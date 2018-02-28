// Command github-team-radar keeps track of your team's work each week in an issue.
package main

import (
	"context"
	"flag"
	"fmt"
	"log"

	"github.com/parkr/github-utils/gh"
	"github.com/parkr/github-utils/radar"
)

func main() {
	var createIssue bool
	flag.BoolVar(&createIssue, "create", false, "Post the issue to GitHub")
	var mention string
	flag.StringVar(&mention, "mention", "", "The user or team to mention at the top of the radar issue")
	var owner string
	flag.StringVar(&owner, "owner", "", "The repository owner the radar issue should be written to")
	var repo string
	flag.StringVar(&repo, "repo", "", "The repository owner the radar issue should be written to")
	flag.Parse()

	if mention == "" {
		log.Fatalln("The -mention flag is required")
	}

	if owner == "" {
		log.Fatalln("The -owner flag is required")
	}

	if repo == "" {
		log.Fatalln("The -repo flag is required")
	}

	client, err := gh.NewDefaultClient()
	if err != nil {
		log.Fatalf("fatal: could not initialize client: %v", err)
	}

	r := radar.NewRadar(&radar.RadarConfig{
		GitHub:    client,
		Mention:   mention,
		RepoOwner: owner,
		RepoName:  repo,
	})

	if createIssue {
		// Create new issue
		newIssue, err := r.Create(context.Background())
		if err != nil {
			log.Fatalf("fatal: error creating radar issue %#v", err)
		}
		log.Println(newIssue.GetHTMLURL())

		// Close the previous issue
		err = r.ClosePrevious(context.Background())
		if err != nil {
			log.Fatalf("fatal: error closing previous radar issue %#v", err)
		}
	} else {
		fmt.Printf("# %s\n\n\n%s\n", r.Title(), r.Body(context.Background()))
	}

}
