package pulls

import (
	"log"
	"strings"

	"github.com/laochailan/barely/maildir"

	"github.com/parkr/github-utils/gh"
	"github.com/shurcooL/githubv4"
)

type User struct {
	Login githubv4.String
	Name  *githubv4.String
}

type Bot struct {
	Login githubv4.String
}

type Author struct {
	User `graphql:"... on User"`
	Bot  `graphql:"... on Bot"`
}

type PullRequest struct {
	Id            githubv4.String
	Number        githubv4.Int
	Title         githubv4.String
	Body          githubv4.String
	Url           githubv4.String
	CreatedAt     githubv4.DateTime
	Author        Author
	TimelineItems struct {
		TotalCount githubv4.Int
	} `graphql:"timelineItems(first: 0)"`
}

type Issue struct {
}

type TimelineEvent struct {
	Body      githubv4.String
	CreatedAt githubv4.DateTime
	Author    Author
}

func FetchPullRequests(client *gh.Client, repo string, input chan PullRequest) error {
	pieces := strings.SplitN(repo, "/", -1)
	owner, name := pieces[0], pieces[1]

	var query struct {
		Repository struct {
			PullRequests struct {
				Edges []struct {
					Cursor *githubv4.String
					Node   PullRequest
				}
			} `graphql:"pullRequests(after: $prCursor, first: 100, states: [CLOSED, MERGED, OPEN])"`
		} `graphql:"repository(owner: $owner, name: $name)"`
	}
	variables := map[string]interface{}{
		"owner":    githubv4.String(owner),
		"name":     githubv4.String(name),
		"prCursor": (*githubv4.String)(nil),
	}

	for {
		err := client.V4.Query(client.Context, &query, variables)
		if err != nil {
			log.Printf("error fetching PR's for '%s': %+v", repo, err)
			close(input)
			return err
		}

		for _, edge := range query.Repository.PullRequests.Edges {
			input <- edge.Node
		}

		if len(query.Repository.PullRequests.Edges) == 0 {
			log.Printf("repo(%s): all done! closing up input...", repo)
			close(input)
			break
		}
		log.Printf("%d", len(query.Repository.PullRequests.Edges))
		lastCursor := query.Repository.PullRequests.Edges[len(query.Repository.PullRequests.Edges)-1].Cursor
		if lastCursor == nil {
			log.Printf("repo(%s): all done! closing up input...", repo)
			close(input)
			break
		}
		variables["prCursor"] = lastCursor
	}

	return nil
}

func CachePullRequestsLocally(client *gh.Client, mailbox *maildir.Dir, repo string, input chan []PullRequest, output chan OfflineStatusResponse) {
	bridge := make(chan OfflineStatusResponse, 1000)
	counter := 0

	//
	for prs := range input {
		prs := prs
		counter += len(prs)
		go func(prs []PullRequest, bridge chan OfflineStatusResponse) {
			for _, pr := range prs {
				bridge <- WritePullRequest(client, mailbox, repo, pr)
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

func GetPullRequestComments(client *gh.Client, owner, repoName string, number int) ([]TimelineEvent, error) {
	return nil, nil
}
