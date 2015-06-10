package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"sync"

	"golang.org/x/oauth2"

	"github.com/google/go-github/github"
	gh "github.com/mfojtik/go-issues/pkg/github"
	"github.com/mfojtik/go-issues/pkg/templates"
)

const (
	// TODO: Make this configurable
	defaultRepo       = "openshift/origin"
	defaultServerAddr = "localhost:8666"

	// TODO: Make this configurable
	cacheFile = "issues-cache.json"
)

// TODO: Make this configurable
var (
	defaultUsers = []string{
		"soltysh",
		"mfojtik",
		"bparees",
		"jhadvig",
		"kargakis",
		"mnagy",
		"csrwng",
		"jcantrill",
		"rhcarvalho",
		"gabemontero",
		"ewolinetz ",
	}
	cli = CLI{}
)

type IssuesList struct {
	Issues map[string][]github.Issue `json:"issues"`
	sync.Mutex
}

type CLI struct {
	UseCache bool
	BindAddr string
}

func init() {
	flag.BoolVar(&cli.UseCache, "use-cache", false, fmt.Sprintf("Attempt to use issues cached in %q", cacheFile))
	flag.StringVar(&cli.BindAddr, "bind", defaultServerAddr, "Bind on this address and port")
}

func WriteIssuesCache(list IssuesList) {
	cacheContent, err := json.Marshal(list)
	if err != nil {
		log.Printf("ERROR: Unable to marshal JSON: %v", err)
		return
	}
	if err := ioutil.WriteFile(cacheFile, []byte(cacheContent), 0644); err != nil {
		log.Printf("ERROR: Unable to write %s: %v", cacheFile, err)
	}
}

func ReadIssuesFromCache() *IssuesList {
	cacheContent, err := ioutil.ReadFile(cacheFile)
	if err != nil {
		log.Printf("Error reading: %v", err)
		return nil
	}
	result := IssuesList{}
	if err := json.Unmarshal(cacheContent, &result); err != nil {
		log.Printf("Error unmarshaling: %v", err)
		return nil
	}
	return &result
}

func main() {
	flag.Parse()

	// Make this work in OpenShift v2
	if len(os.Getenv("HOST")) > 0 && len(os.Getenv("PORT")) > 0 {
		cli.BindAddr = fmt.Sprintf("%s:%s", os.Getenv("HOST"), os.Getenv("PORT"))
	}

	var (
		issues   IssuesList
		err      error
		hasCache bool
	)
	token := os.Getenv("GITHUB_AUTH_TOKEN")
	tc := oauth2.NewClient(oauth2.NoContext, oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
	))
	client := github.NewClient(tc)

	if cli.UseCache {
		if cachedIssues := ReadIssuesFromCache(); cachedIssues != nil {
			issues = *cachedIssues
			hasCache = true
			log.Printf("Using %d issues found in %q cache file\n", len(issues.Issues), cacheFile)
		}
	}

	if !hasCache {
		issues.Issues, err = gh.FetchIssues(client, defaultRepo, defaultUsers)
		if err != nil {
			log.Printf("Error fetching issues from Github: %v\n", err)
		}
	}

	if cli.UseCache {
		log.Printf("Caching issues list to %q\n", cacheFile)
		WriteIssuesCache(issues)
	}

	log.Printf("Starting HTTP server %q ...\n", "http://"+cli.BindAddr)
	http.HandleFunc("/", func(w http.ResponseWriter, req *http.Request) {
		if req.URL.Path != "/" {
			http.NotFound(w, req)
			return
		}
		if len(req.URL.Query().Get("refresh")) > 0 {
			issues.Lock()
			issues.Issues, err = gh.FetchIssues(client, defaultRepo, defaultUsers)
			issues.Unlock()
			if err != nil {
				fmt.Fprintf(w, "ERROR: %v", err)
				return
			}
		}
		t := template.New("issues")
		t, _ = t.Parse(templates.IssuesTemplate)
		t.Execute(w, issues)
		defer func() {
			if r := recover(); r != nil {
				fmt.Println("Recovered from panic!", r)
			}
		}()
	})

	log.Fatal(http.ListenAndServe(cli.BindAddr, nil))
}
