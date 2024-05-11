// A command-line utility to run a search query and display results.
package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"

	"github.com/google/go-github/v62/github"
	"github.com/parkr/github-utils/gh"
	"github.com/parkr/github-utils/search"
	"github.com/parkr/github-utils/webview"
)

func haltIfError(err error) {
	if err != nil {
		log.Fatalln("error:", err)
	}
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
	issues, err := search.FindAllAssignedIssues(client, user)
	haltIfError(err)
	printIssues(fmt.Sprintf("issues assigned to %s", user), issues)

	issues, err = search.FindAllUnansweredMentions(client, user)
	haltIfError(err)
	printIssues(fmt.Sprintf("unanswered issues for %s", user), issues)
}

func main() {
	var httpBind string
	flag.StringVar(&httpBind, "http", "", "The network binding to attach a server to. Only boots server if specified.")
	flag.Parse()

	client, err := gh.NewDefaultClient()
	if err != nil {
		log.Fatalf("fatal: could not initialize client: %v", err)
	}

	users := []string{"parkr"}
	if flag.NArg() > 0 {
		users = flag.Args()
	}

	if httpBind != "" {
		handler := webview.Handler{Users: users, Client: client}
		http.Handle("/", handler)
		log.Println("Starting server on", httpBind)
		if err := http.ListenAndServe(httpBind, nil); err != nil {
			log.Fatalln(err)
		}
	} else {
		// Print to stdout.
		for _, user := range users {
			unearthForUser(client, user)
		}
	}
}
