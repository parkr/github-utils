package search

import (
	"fmt"
	"log"
	"strings"

	"github.com/google/go-github/github"
	"github.com/parkr/github-utils/gh"
)

func SearchIssues(client *gh.Client, query string) ([]github.Issue, error) {
	results := []github.Issue{}

	input := make(chan github.Issue, 100)
	queue := make(chan github.Issue, 100)
	go prefillSearchResultIssues(query, input, queue)

	opts := &github.SearchOptions{
		Sort:        "created",
		Order:       "asc",
		ListOptions: github.ListOptions{PerPage: 100},
	}
	for {
		searchResult, resp, err := client.Search.Issues(query, opts)
		if err != nil {
			log.Printf("error issuing query '%s': %+v", query, err)
			close(input)
			return nil, err
		}

		for _, issue := range searchResult.Issues {
			issue := issue
			input <- issue
		}

		if resp.NextPage == 0 {
			log.Printf("query(%s): all done! closing up input...", query)
			close(input)
			break
		}
		opts.ListOptions.Page = resp.NextPage
	}

	log.Printf("query(%s): waiting for queue...", query)
	for issue := range queue {
		results = append(results, issue)
	}
	log.Printf("query(%s): queue finished!", query)

	return results, nil
}

func prefillSearchResultIssues(query string, input chan github.Issue, queue chan github.Issue) {
	log.Printf("query(%s): waiting for input...", query)
	for issue := range input {
		log.Printf("query(%s): processing %s", query, *issue.HTMLURL)
		if issue.Repository == nil {
			pieces := strings.SplitN(*issue.URL, "/", -1)
			owner, repo := pieces[4], pieces[5]
			issue.Repository = &github.Repository{
				Owner:    &github.User{Login: github.String(owner)},
				Name:     github.String(repo),
				FullName: github.String(owner + "/" + repo),
				HTMLURL:  github.String("https://github.com/" + owner + "/" + repo),
			}
		}
		queue <- issue
	}
	log.Printf("query(%s): input finished!", query)
	close(queue)
}

func FindAllUnansweredMentions(client *gh.Client, user string) ([]github.Issue, error) {
	query := fmt.Sprintf("is:open mentions:%s -commenter:%s", user, user)
	return SearchIssues(client, query)
}

func FindAllAssignedIssues(client *gh.Client, user string) ([]github.Issue, error) {
	query := fmt.Sprintf("is:open assignee:%s", user)
	return SearchIssues(client, query)
}

func FindAllCreatedIssues(client *gh.Client, user string) ([]github.Issue, error) {
	query := fmt.Sprintf("is:open author:%s", user)
	return SearchIssues(client, query)
}
