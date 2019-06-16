package main

import (
	"flag"
	"fmt"
	"log"

	"github.com/parkr/github-utils/gh"
	"github.com/parkr/github-utils/langclassify"
	"github.com/parkr/github-utils/repos"
)

func main() {
	user := flag.String("user", "octocat", "GitHub Username to list repos for")
	flag.Parse()

	client, err := gh.NewDefaultClient()
	if err != nil {
		log.Fatalf("fatal: could not initialize client: %v", err)
	}

	userRepos, err := repos.AllReposForUser(client, *user)
	if err != nil {
		log.Fatalf("fatal: could not fetch repos for %s: %+v", *user, err)
	}

	log.Printf("user(%s): fetched %d repositories", *user, len(userRepos))

	reposByLanguage := map[string][]string{}

	for _, repo := range userRepos {
		lang, err := langclassify.ClassifyRepository(client, repo)
		if err != nil {
			log.Printf("repo(%s): unable to classify repository: %+v", repo.GetFullName(), err)
			continue
		}
		if _, ok := reposByLanguage[lang.GetName()]; !ok {
			reposByLanguage[lang.GetName()] = []string{}
		}
		reposByLanguage[lang.GetName()] = append(reposByLanguage[lang.GetName()],
			fmt.Sprintf("%s\t%s", repo.GetFullName(), lang.GetVersion(client)),
		)
	}

	for lang, langRepos := range reposByLanguage {
		log.Printf("%s:", lang)
		for _, repo := range langRepos {
			log.Printf("\t%s", repo)
		}
	}
}
