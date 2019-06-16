package main

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/google/go-github/github"
	"github.com/parkr/github-utils/gh"
)

func main() {
	client, err := gh.NewDefaultClient()
	if err != nil {
		log.Fatalf("fatal: could not initialize client: %v", err)
	}

	allWatchedRepos := []*github.Repository{}
	listOpts := &github.ListOptions{PerPage: 200}
	start := time.Now()

	for {
		repos, resp, err := client.Activity.ListWatched(context.Background(), "", listOpts)
		if err != nil {
			log.Fatalf("error listing watched repos: %+v", err)
		}
		allWatchedRepos = append(allWatchedRepos, repos...)

		log.Printf("[debug] fetched %d repos, next page is %d", len(repos), resp.NextPage)

		if resp.NextPage == 0 {
			break
		}
		listOpts.Page = resp.NextPage
	}
	log.Printf("[debug] fetched watching repos in %s", time.Since(start))

	log.Printf("You are watching %d repositories.", len(allWatchedRepos))

	reader := bufio.NewReader(os.Stdin)
	for _, repo := range allWatchedRepos {
		log.Printf("Would you like to unwatch %s?", repo.GetFullName())
		log.Println("Description:", repo.GetDescription())
		fmt.Printf("(y/n) --> ")
		text, err := reader.ReadString('\n')
		if err != nil {
			log.Fatalf("Oops, I didn't quite catch that. Error: %+v", err)
		}
		if text == "y" || text == "y\n" || text == "yes" || text == "yes\n" {
			_, err := client.Activity.DeleteRepositorySubscription(
				context.Background(),
				repo.GetOwner().GetLogin(),
				repo.GetName(),
			)
			if err == nil {
				log.Printf("Unwatched %s.", repo.GetFullName())
			} else {
				log.Printf("Oops, couldn't unwatch %s: %+v", repo.GetFullName(), err)
			}
		} else {
			log.Printf("Still watching %s.", repo.GetFullName())
		}
	}
}
