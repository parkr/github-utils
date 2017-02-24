package pulls

import (
	"log"
	"strings"

	"github.com/google/go-github/github"
	"github.com/parkr/github-utils/gh"
)

func FetchPullRequests(client *gh.Client, repo string, input chan *github.PullRequest) error {
	pieces := strings.SplitN(repo, "/", -1)
	owner, name := pieces[0], pieces[1]

	opts := &github.PullRequestListOptions{
		State:       "open",
		Sort:        "created",
		ListOptions: github.ListOptions{PerPage: 100},
	}

	for {
		prs, resp, err := client.PullRequests.List(client.Context, owner, name, opts)
		if err != nil {
			log.Printf("error fetching PR's for '%s': %+v", repo, err)
			close(input)
			return err
		}

		for _, pr := range prs {
			input <- pr
		}

		if resp.NextPage == 0 {
			log.Printf("repo(%s): all done! closing up input...", repo)
			close(input)
			break
		}
		opts.ListOptions.Page = resp.NextPage
	}

	return nil
}

func CachePullRequestsLocally(client *gh.Client, outputDir string, input chan []*github.PullRequest, output chan OfflineStatusResponse) {
	bridge := make(chan OfflineStatusResponse, 1000)
	counter := 0

	//
	for prs := range input {
		prs := prs
		counter += len(prs)
		go func(prs []*github.PullRequest, bridge chan OfflineStatusResponse) {
			for _, pr := range prs {
				bridge <- WritePullRequest(client, outputDir, pr)
			}
		}(prs, bridge)
	}

	log.Printf("Saving %d pull requests...", counter)

	// Now wait for all of them to process.
	for status := range bridge {
		counter--
		output <- status
		if counter == 0 {
			break
		}
	}
	close(bridge)
	close(output)
}
