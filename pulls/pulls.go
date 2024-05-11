package pulls

import (
	"log"
	"strings"

	"github.com/google/go-github/v62/github"
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

func CachePullRequestsLocally(client *gh.Client, outputDir, repo string, input chan []*github.PullRequest, output chan OfflineStatusResponse) {
	bridge := make(chan OfflineStatusResponse, 1000)
	counter := 0

	//
	for prs := range input {
		prs := prs
		counter += len(prs)
		go func(prs []*github.PullRequest, bridge chan OfflineStatusResponse) {
			for _, pr := range prs {
				bridge <- WritePullRequest(client, outputDir, repo, pr)
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

func GetPullRequestComments(client *gh.Client, owner, repoName string, number int) ([]*github.IssueComment, error) {
	comments, nwo := []*github.IssueComment{}, owner+"/"+repoName
	opts := &github.IssueListCommentsOptions{
		Sort:        github.String("created"),
		Direction:   github.String("asc"),
		ListOptions: github.ListOptions{PerPage: 100},
	}
	for {
		apiComments, resp, err := client.Issues.ListComments(client.Context, owner, repoName, number, opts)
		if err != nil {
			log.Printf("error fetching PR line comments for '%s': %+v", nwo, err)
			return nil, err
		}

		comments = append(comments, apiComments...)

		if resp.NextPage == 0 {
			break
		}
		opts.ListOptions.Page = resp.NextPage
	}
	return comments, nil
}

func GetPullRequestLineComments(client *gh.Client, owner, repoName string, number int) ([]*github.PullRequestComment, error) {
	comments, nwo := []*github.PullRequestComment{}, owner+"/"+repoName
	opts := &github.PullRequestListCommentsOptions{
		Sort:        "created",
		Direction:   "asc",
		ListOptions: github.ListOptions{PerPage: 100},
	}
	for {
		apiComments, resp, err := client.PullRequests.ListComments(client.Context, owner, repoName, number, opts)
		if err != nil {
			log.Printf("error fetching PR line comments for '%s': %+v", nwo, err)
			return nil, err
		}

		comments = append(comments, apiComments...)

		if resp.NextPage == 0 {
			break
		}
		opts.ListOptions.Page = resp.NextPage
	}
	return comments, nil
}
