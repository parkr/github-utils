package repos

import (
	"log"

	"github.com/google/go-github/github"
	"github.com/parkr/github-utils/gh"
)

type Repository struct {
	*github.Repository
}

// AllReposForUser returns all the repositories for the user.
func AllReposForUser(client *gh.Client, user string) ([]*Repository, error) {
	opts := &github.RepositoryListOptions{
		Type:        "all",       // valid options: all, owner, member. default: owner
		Sort:        "full_name", // valid options: created, updated, pushed, full_name. default: full_name
		ListOptions: github.ListOptions{Page: 0, PerPage: 100},
	}

	allRepos := []*Repository{}

	for {
		log.Printf("user(%s): fetching page %d of repositories", user, opts.ListOptions.Page)
		repos, resp, err := client.Repositories.List(client.Context, user, opts)
		if err != nil {
			log.Printf("error fetching PR's for user %q: %+v", user, err)
			return nil, err
		}

		for _, repo := range repos {
			if *repo.Fork {
				continue // skip forks
			}
			allRepos = append(allRepos, &Repository{repo})
		}

		if resp.NextPage == 0 {
			break
		}
		opts.ListOptions.Page = resp.NextPage
	}

	return allRepos, nil
}
