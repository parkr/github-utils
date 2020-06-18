package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"time"

	"github.com/google/go-github/v28/github"
	"github.com/parkr/github-utils/gh"
)

func processRepos(client *gh.Client, repos []*github.Repository, newDefaultBranchName string) {
	ctx := context.Background()

	for _, repo := range repos {
		if repo.GetDefaultBranch() != newDefaultBranchName {
			if repo.Description != nil {
				fmt.Printf("%s - %s\n", *repo.FullName, *repo.Description)
			} else {
				fmt.Printf("%s\n", *repo.FullName)
			}
			fmt.Printf("  change default branch to %q? (y/n) > \n", newDefaultBranchName)
			response := ""
			_, err := fmt.Scanln(&response)
			if err != nil {
				log.Fatalln(err)
			}
			if response == "y" {
				// Create branch if it doesn't exist
				if _, _, err := client.Repositories.GetBranch(ctx, *repo.Owner.Login, *repo.Name, newDefaultBranchName); err != nil {
					// We got an error, so we should create the branch.
					_, _, _, := client.Repositories.CreateBranch(ctx, *repo.Owner.Login, *repo.Name, newDefaultBranchName)
					if err != nil {
						log.Printf("error creating branch: %v", err)
					} else {
						fmt.Println("  ... done")
					}
				}
				_, _, err := client.Repositories.Edit(ctx, *repo.Owner.Login, *repo.Name, &github.Repository{
					DefaultBranch: github.String(newDefaultBranchName),
				})
				if err != nil {
					log.Printf("error updating default branch: %v", err)
				} else {
					fmt.Println("  ... done")
				}
			} else {
				fmt.Println("  ... skipped")
			}
		}
	}
}

func main() {
	newDefaultBranchName := flag.String("new-name", "main", "The new name to use for the default branch on given repos")
	flag.Parse()

	client, err := gh.NewDefaultClient()
	if err != nil {
		log.Fatalf("fatal: could not initialize client: %v", err)
	}

	ctx, cancel := context.WithTimeout(client.Context, 5*time.Minute)
	defer cancel()

	opt := &github.RepositoryListOptions{
		ListOptions: github.ListOptions{PerPage: 100},
	}
	for {
		repos, resp, err := client.Repositories.List(ctx, client.Login, opt)
		if err != nil {
			log.Fatalf("fatal: %v", err)
		}
		processRepos(client, repos, *newDefaultBranchName)
		if resp.NextPage == 0 {
			break
		}
		opt.ListOptions.Page = resp.NextPage
	}
}
