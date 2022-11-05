// A command to process one or all of a user's repos to see if Dependabot has
// been setup and that it covers all ecosystems in the repository.
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

func listAllRepos(ctx context.Context, client *gh.Client, githubLogin string, repoChan chan string, done chan bool) {
	// If org, use one method. If user, use another.
	githubUser, _, err := client.Users.Get(ctx, githubLogin)
	if err != nil {
		log.Fatalf("fatal: unable to get user %q: %+v", githubLogin, err)
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

	log.Println("listing repos for", githubLogin)

	opt := &github.ListOptions{PerPage: 100}
	for {
		repos, resp, err := listMethod(ctx, githubLogin, opt)
		if err != nil {
			log.Fatalf("fatal: %v", err)
		}
		if resp.StatusCode != 200 {
			log.Fatalf("fatal: unexpected status: %s", resp.Status)
		}
		if len(repos) == 0 {
			fmt.Printf("fatal: no repos for github login %q", githubLogin)
			break
		}
		for _, repo := range repos {
			repo := repo
			if repo.GetArchived() {
				log.Printf("[%s]: archived, skipping", repo.GetFullName())
				continue
			}
			if repo.GetFork() {
				log.Printf("[%s]: fork, skipping", repo.GetFullName())
				continue
			}
			// log.Printf("[%s] enqueueing", repo.GetFullName())
			repoChan <- repo.GetName()
		}
		if resp.NextPage == 0 {
			break
		}
		opt.Page = resp.NextPage
	}
	close(repoChan)
	done <- true
}

func dependabotAuditForSingleRepo(ctx context.Context, client *gh.Client, githubLogin string, repoName string) {
	// What dependabot ecosystems should be declared?
	checkedEcosystems := map[string]bool{}
	// 1. Go: go.mod file
	checkedEcosystems["gomod"] = containsFile(ctx, client, githubLogin, repoName, "go.mod")
	// 2. Python: requirements.txt file
	checkedEcosystems["pip"] = containsFile(ctx, client, githubLogin, repoName, "requirements.txt")
	// 3. Docker: Dockerfile
	checkedEcosystems["docker"] = containsFile(ctx, client, githubLogin, repoName, "Dockerfile")
	// 4. Ruby: Gemfile
	checkedEcosystems["bundler"] = containsFile(ctx, client, githubLogin, repoName, "Gemfile")
	// 5. Actions: .github/workflows directory
	checkedEcosystems["github-actions"] = containsGitHubActions(ctx, client, githubLogin, repoName)
	// 6. Node: package.json file
	checkedEcosystems["npm"] = containsFile(ctx, client, githubLogin, repoName, "package.json")
	// 7. Rust: Cargofile
	checkedEcosystems["cargo"] = containsFile(ctx, client, githubLogin, repoName, "Cargo.toml")
	// log.Printf("[%s/%s]: checked ecosystems: %#v", githubLogin, repoName, checkedEcosystems)

	// No ecosystems? Quickly return.
	if len(checkedEcosystems) == 0 {
		log.Printf("[%s/%s]: no supported updateable files found", githubLogin, repoName)
		return
	}

	// Read ecosystems that are currently configured.
	currentlyConfiguredEcosystems := map[string]bool{}
	dependabotFile := readDependabotConfig(ctx, client, githubLogin, repoName)
	for _, dependabotUpdateConfig := range dependabotFile.Updates {
		currentlyConfiguredEcosystems[dependabotUpdateConfig.PackageEcosystem] = true
	}
	// log.Printf("[%s/%s]: currently configured: %#v", githubLogin, repoName, currentlyConfiguredEcosystems)

	// Compare what should be declared and what is declared.
	for ecosystem, isRequired := range checkedEcosystems {
		if isRequired && !currentlyConfiguredEcosystems[ecosystem] {
			log.Printf("[%s/%s]: missing ecosystem: %s", githubLogin, repoName, ecosystem)
		}
	}
}

func dependabotAuditForRepos(ctx context.Context, client *gh.Client, githubLogin string, repoChan chan string, done chan bool) {
	for {
		select {
		case <-ctx.Done():
			if ctx.Err() != nil {
				log.Fatalln(ctx.Err())
			}
			return
		case repoName, more := <-repoChan:
			if !more {
				done <- true
				return
			}
			dependabotAuditForSingleRepo(ctx, client, githubLogin, repoName)
		}
	}
}

func main() {
	githubLogin := flag.String("login", "", "GitHub Login (user or org) whose repos to list (default: currently-authorized user)")
	singleRepo := flag.String("repo", "", "Single repo to audit (default: audit all repos for the login")
	flag.Parse()

	if githubLogin == nil || *githubLogin == "" {
		log.Fatalln("fatal: -login flag required")
	}

	client, err := gh.NewDefaultClient()
	if err != nil {
		log.Fatalf("fatal: could not initialize client: %v", err)
	}

	ctx, cancel := context.WithTimeout(client.Context, 5*time.Minute)
	defer cancel()

	done := make(chan bool, 2)
	repoChan := make(chan string, 10)

	go dependabotAuditForRepos(ctx, client, *githubLogin, repoChan, done)
	go dependabotAuditForRepos(ctx, client, *githubLogin, repoChan, done)
	go dependabotAuditForRepos(ctx, client, *githubLogin, repoChan, done)

	if singleRepo != nil && *singleRepo != "" {
		repoChan <- *singleRepo
		close(repoChan)
		done <- true
	} else {
		listAllRepos(ctx, client, *githubLogin, repoChan, done)
	}

	<-done // listAllRepos
	<-done // actionsAuditForRepos
	<-done // actionsAuditForRepos 2
	<-done // actionsAuditForRepos 3
	log.Println("audit complete")
}
