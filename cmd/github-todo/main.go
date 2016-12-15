// A command-line utility to run a search query and display results.
package main

import (
	"flag"
	"fmt"
	"log"
	"strings"

	"github.com/google/go-github/github"
	"github.com/parkr/github-utils/gh"
)

func haltIfError(err error) {
	if err != nil {
		log.Fatalln("error:", err)
	}
}

func repoNameFromURL(url string) string {
	return strings.Join(
		strings.SplitN(
			strings.Replace(url, "https://github.com/", "", 1),
			"/",
			-1)[1:2],
		"/",
	)
}

func issuesForQuery(client *gh.Client, query string) []github.Issue {
	results := []github.Issue{}

	opts := &github.SearchOptions{
		Sort:        "created",
		ListOptions: github.ListOptions{PerPage: 100},
	}
	for {
		searchResult, resp, err := client.Search.Issues(query, opts)
		haltIfError(err)

		// Append the results.
		results = append(results, searchResult.Issues...)

		if resp.NextPage == 0 {
			break
		}
		opts.ListOptions.Page = resp.NextPage
	}

	return results
}

func findAllUnansweredMentions(client *gh.Client, user string) []github.Issue {
	query := fmt.Sprintf("is:open mentions:%s -commenter:%s", user, user)
	return issuesForQuery(client, query)
}

func truncate(text string, max int) string {
	if len(text) <= max {
		return text
	}

	return text[0:max-3] + "..."
}

func printIssues(title string, issues []github.Issue) {
	fmt.Printf("%d %s:\n", len(issues), title)
	for _, issue := range issues {
		fmt.Printf("%s | %-50s | %s\n",
			issue.CreatedAt.Format("2006-01-02"),
			truncate(*issue.Title, 50),
			*issue.HTMLURL,
		)
	}
}

func unearthForUser(client *gh.Client, user string) {
	// issues := findAllAssignedIssues(client, user)
	// printIssues("assigned issues", issues)

	issues := findAllUnansweredMentions(client, user)
	printIssues("unanswered issues", issues)
}

func main() {
	flag.Parse()
	client, err := gh.NewDefaultClient()
	if err != nil {
		log.Fatalf("fatal: could not initialize client: %v", err)
	}

	users := []string{"parkr"}
	if flag.NArg() > 0 {
		users = flag.Args()
	}

	for _, user := range users {
		unearthForUser(client, user)
	}
}
