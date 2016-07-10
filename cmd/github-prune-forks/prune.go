package main

import (
	"fmt"
	"log"

	"github.com/google/go-github/github"
	"github.com/parkr/github-utils/gh"
)

func processRepos(client *gh.Client, repos []*github.Repository) {
	for _, repo := range repos {
		if *repo.Fork {
			if repo.Description != nil {
				fmt.Printf("%s - %s\n", *repo.FullName, *repo.Description)
			} else {
				fmt.Printf("%s\n", *repo.FullName)
			}
			fmt.Print("  remove? (y/n) > ")
			response := ""
			_, err := fmt.Scanln(&response)
			if err != nil {
				log.Fatalln(err)
			}
			if response == "y" {
				_, err := client.Repositories.Delete(*repo.Owner.Login, *repo.Name)
				if err != nil {
					log.Printf("error: %v", err)
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
	client, err := gh.NewDefaultClient()
	if err != nil {
		log.Fatalf("fatal: could not initialize client: %v", err)
	}

	opt := &github.RepositoryListOptions{
		ListOptions: github.ListOptions{PerPage: 100},
	}
	for {
		repos, resp, err := client.Repositories.List(client.Login, opt)
		if err != nil {
			log.Fatalf("fatal: %v", err)
		}
		processRepos(client, repos)
		if resp.NextPage == 0 {
			break
		}
		opt.ListOptions.Page = resp.NextPage
	}
}
