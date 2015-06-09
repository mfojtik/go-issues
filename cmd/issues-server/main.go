package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"sync"

	"github.com/google/go-github/github"
	"github.com/mfojtik/go-issues/pkg/templates"
)

const (
	OrgName    = "openshift"
	RepoName   = "origin"
	serverAddr = "localhost:8666"
	cacheFile  = "issues-cache.json"
)

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
	}
	cli = CLI{}
)

type IssuesList struct {
	Issues map[string][]github.Issue `json:"issues"`
	sync.Mutex
}

type CLI struct {
	UseCache bool
}

func init() {
	flag.BoolVar(&cli.UseCache, "use-cache", false, fmt.Sprintf("Attempt to use issues cached in %q", cacheFile))
}

func FetchIssuesForUser(user string) []github.Issue {
	client := github.NewClient(nil)
	issues, _, err := client.Issues.ListByRepo(OrgName, RepoName, &github.IssueListByRepoOptions{
		Assignee: user,
	})
	if err != nil {
		log.Fatalf("Unable to list issues: %v", err)
	}
	return issues
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

	var (
		issues   IssuesList
		hasCache bool
		wg       sync.WaitGroup
	)

	if cli.UseCache {
		if cachedIssues := ReadIssuesFromCache(); cachedIssues != nil {
			issues.Lock()
			issues = *cachedIssues
			issues.Unlock()
			hasCache = true
			log.Printf("Using %d issues found in %q cache file\n", len(issues.Issues), cacheFile)
		}
	}

	if !hasCache {
		issues.Issues = make(map[string][]github.Issue)
		for _, user := range defaultUsers {
			wg.Add(1)
			go func(u string) {
				issues.Lock()
				defer wg.Done()
				defer issues.Unlock()
				issues.Issues[u] = FetchIssuesForUser(u)
				log.Printf("Fetched %d issues assigned to @%s\n", len(issues.Issues[u]), u)
			}(user)
		}
		log.Printf("Fetching GitHub issues ...\n")
		wg.Wait()
	}

	if cli.UseCache {
		log.Printf("Caching issues list to %q\n", cacheFile)
		WriteIssuesCache(issues)
	}

	log.Printf("Starting HTTP server %q ...\n", "http://"+serverAddr)
	http.HandleFunc("/", func(w http.ResponseWriter, req *http.Request) {
		if req.URL.Path != "/" {
			http.NotFound(w, req)
			return
		}
		t := template.New("issues")
		t, _ = t.Parse(templates.IssuesTemplate)
		issues.Lock()
		t.Execute(w, issues)
		defer func() {
			issues.Unlock()
			if r := recover(); r != nil {
				fmt.Println("Recovered from panic!", r)
			}
		}()
	})

	log.Fatal(http.ListenAndServe(serverAddr, nil))
}
