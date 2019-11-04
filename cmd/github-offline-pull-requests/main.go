// A command-line utility to download pull requests for offline reading.
package main

import (
	"flag"
	"log"
	"os"

	"github.com/laochailan/barely/maildir"

	"github.com/parkr/github-utils/gh"
	"github.com/parkr/github-utils/pulls"
)

// 1. Read in pull requests from the API. Push into queue. Close queue when finished.
// 2. Read from queue, and batch process 10 pull requests at once:
//     2a. Write a `.patch` file for the pull request.
//     2b. Write a metadata file, preferably in Markdown or something human-readable.
//           - Username of author
//           - PR reviews & comments
//           - PR comment chain

func writeOutputDirectory(dir string) error {
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		return os.MkdirAll(dir, 0755)
	}
	return nil
}

func batchPullRequests(input chan pulls.PullRequest, bridge chan []pulls.PullRequest) {
	counter := 0
	prs := []pulls.PullRequest{}
	for pr := range input {
		counter++
		prs = append(prs, pr)
		if counter%5 == 0 {
			bridge <- prs
			prs = []pulls.PullRequest{}
		}
	}
	if len(prs) > 0 {
		bridge <- prs
	}
	close(bridge)
}

func main() {
	cwd, err := os.Getwd()
	if err != nil {
		log.Fatalf("fatal: could not get $CWD: %+v", err)
	}

	var repo string
	flag.StringVar(&repo, "repo", "", "The repository NWO (e.g. parkr/auto-reply) to copy locally.")
	var dir string
	flag.StringVar(&dir, "dir", cwd, "Output directory, defaults to $CWD.")
	flag.Parse()

	if repo == "" {
		log.Fatalln("fatal: missing -repo")
	}

	client, err := gh.NewDefaultClient()
	if err != nil {
		log.Fatalf("fatal: could not initialize client: %v", err)
	}

	if err := writeOutputDirectory(dir); err != nil {
		log.Fatalf("fatal: could not create maildir %q: %v", dir, err)
	}

	mailbox, err := maildir.Open(dir, true)
	if err != nil {
		log.Fatalf("fatal: could not create maildir %q: %v", dir, err)
	}

	input := make(chan pulls.PullRequest, 100)
	bridge := make(chan []pulls.PullRequest)
	output := make(chan pulls.OfflineStatusResponse)

	go pulls.FetchPullRequests(client, repo, input)
	go batchPullRequests(input, bridge)
	go pulls.CachePullRequestsLocally(client, mailbox, repo, bridge, output)

	for resp := range output {
		if resp.Success {
			log.Printf("Wrote %s#%d to %s", repo, resp.Number, resp.Filename)
		} else {
			log.Printf("Fetching PR %s#%d failed: %+v", repo, resp.Number, resp.Error)
		}
	}
}
