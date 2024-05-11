package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"time"

	"github.com/google/go-github/v62/github"
	"github.com/parkr/github-utils/gh"
)

const maxRedirectsFetchingBranch = 1

func processRepos(client *gh.Client, repos []*github.Repository, newDefaultBranchName string) {
	ctx := context.Background()

	for _, repo := range repos {
		if repo.GetDefaultBranch() == newDefaultBranchName {
			continue
		}

		if repo.GetArchived() {
			continue
		}

		if repo.Description != nil {
			fmt.Printf("%s - %s\n", *repo.FullName, *repo.Description)
		} else {
			fmt.Printf("%s\n", *repo.FullName)
		}
		fmt.Printf("  %s\n", repo.GetHTMLURL())
		fmt.Printf("  change default branch from %q to %q? (y/n) > ", repo.GetDefaultBranch(), newDefaultBranchName)
		response := ""
		_, err := fmt.Scanln(&response)
		if err != nil {
			log.Fatalln(err)
		}
		if response == "y" {
			// Create branch if it doesn't exist
			if _, _, err := client.Repositories.GetBranch(ctx, *repo.Owner.Login, *repo.Name, newDefaultBranchName, maxRedirectsFetchingBranch); err != nil {
				// We got an error, so we should create the branch.
				oldRef, _, err := client.Git.GetRef(ctx, *repo.Owner.Login, *repo.Name, "refs/heads/"+repo.GetDefaultBranch())
				if err != nil {
					log.Printf("error fetching old ref %q: %+v", repo.GetDefaultBranch(), err)
					continue
				}
				newRef := &github.Reference{
					Ref:    github.String("refs/heads/" + newDefaultBranchName),
					Object: oldRef.Object,
				}
				if _, _, err := client.Git.CreateRef(ctx, *repo.Owner.Login, *repo.Name, newRef); err != nil {
					log.Printf("error creating branch: %v", err)
					continue
				}
			}
			_, _, err := client.Repositories.Edit(ctx, *repo.Owner.Login, *repo.Name, &github.Repository{
				DefaultBranch: github.String(newDefaultBranchName),
			})
			if err != nil {
				log.Printf("error updating default branch: %v", err)
			}
			if _, err := client.Git.DeleteRef(ctx, *repo.Owner.Login, *repo.Name, "refs/heads/"+repo.GetDefaultBranch()); err != nil {
				log.Printf("error deleting old default branch: %v", err)
			}
			fmt.Println("  ... done")
		}
	}
}

func main() {
	newDefaultBranchName := flag.String("new-name", "main", "The new name to use for the default branch on given repos")
	githubLogin := flag.String("login", "", "GitHub Login (user or org) whose repos to list (default: currently-authorized user)")
	flag.Parse()

	client, err := gh.NewDefaultClient()
	if err != nil {
		log.Fatalf("fatal: could not initialize client: %v", err)
	}

	ctx, cancel := context.WithTimeout(client.Context, 5*time.Minute)
	defer cancel()

	// If org, use one method. If user, use another.
	githubUser, _, err := client.Users.Get(ctx, *githubLogin)
	if err != nil {
		log.Fatalf("fatal: unable to get user %q: %+v", *githubLogin, err)
	}

	// There are two different ways to list repos: one by user and one by org.
	var listMethod func(context.Context, string, *github.ListOptions) ([]*github.Repository, *github.Response, error)
	if githubUser.GetType() == "User" {
		listMethod = func(ctx context.Context, login string, opts *github.ListOptions) ([]*github.Repository, *github.Response, error) {
			return client.Repositories.List(ctx, login, &github.RepositoryListOptions{
				Type:        "owner",
				ListOptions: *opts,
			})
		}
	} else {
		listMethod = func(ctx context.Context, login string, opts *github.ListOptions) ([]*github.Repository, *github.Response, error) {
			return client.Repositories.ListByOrg(ctx, login, &github.RepositoryListByOrgOptions{
				Type:        "all",
				ListOptions: *opts,
			})
		}
	}

	opt := &github.ListOptions{PerPage: 100}
	for {
		repos, resp, err := listMethod(ctx, *githubLogin, opt)
		if err != nil {
			log.Fatalf("fatal: %v", err)
		}
		if len(repos) == 0 {
			fmt.Printf("fatal: no repos for github login %q", *githubLogin)
			break
		}
		processRepos(client, repos, *newDefaultBranchName)
		if resp.NextPage == 0 {
			break
		}
		opt.Page = resp.NextPage
	}
}
