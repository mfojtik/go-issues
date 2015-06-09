package github

import (
	"log"
	"strings"
	"sync"

	"github.com/google/go-github/github"
)

func FetchIssuesForUser(client *github.Client, repo, user string) ([]github.Issue, error) {
	repoParts := strings.Split(repo, "/")
	issues, _, err := client.Issues.ListByRepo(repoParts[0], repoParts[1], &github.IssueListByRepoOptions{
		Assignee: user,
	})
	return issues, err
}

func FetchIssues(client *github.Client, repo string, people []string) (map[string][]github.Issue, error) {
	var (
		wg          sync.WaitGroup
		resultError error
	)
	result := make(map[string][]github.Issue)
	for _, person := range people {
		wg.Add(1)
		go func(u string) {
			defer wg.Done()
			if items, err := FetchIssuesForUser(client, repo, u); err != nil {
				resultError = err
				return
			} else {
				result[u] = items
			}
			log.Printf("Fetched %d issues assigned to @%s\n", len(result[u]), u)
		}(person)
	}
	log.Printf("Fetching GitHub issues ...\n")
	wg.Wait()
	return result, resultError
}
