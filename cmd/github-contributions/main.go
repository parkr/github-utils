// A command-line utility to generate a report of a user's contributions on GitHub.
package main

import (
	"flag"
	"log"
	"os"
	"time"

	"github.com/parkr/github-utils/contributions"
	"github.com/parkr/github-utils/gh"
)

var oneWeekAgo = time.Now().AddDate(0, 0, -7).Format("2006-01-02")

func main() {
	var login string
	flag.StringVar(&login, "login", "", "The GitHub username of the user, e.g. 'defunkt'")
	var startDate string
	flag.StringVar(&startDate, "since", oneWeekAgo, "The start date to look for contributions")
	var owner string
	flag.StringVar(&owner, "owner", "", "The owner to which to scope our contribution scopes, e.g. 'github'")
	flag.Parse()

	if login == "" {
		log.Fatal("error: you must specify a -login")
	}

	client, err := gh.NewDefaultClient()
	if err != nil {
		log.Fatalf("fatal: could not initialize client: %v", err)
	}

	err = contributions.New(client, login, startDate, owner).Write(os.Stdout)
	if err != nil {
		log.Fatalf("error: %+v", err)
	}
}
