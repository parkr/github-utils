package main

import (
	"context"
	"log"
	"net/http"

	"github.com/parkr/github-utils/gh"
	"gopkg.in/yaml.v2"
)

func containsGitHubActions(ctx context.Context, client *gh.Client, githubLogin string, repoName string) bool {
	path := ".github/workflows"
	_, directoryContents, resp, err := client.Repositories.GetContents(ctx, githubLogin, repoName, path, nil)
	if resp.StatusCode == http.StatusNotFound {
		return false
	}
	if err != nil {
		log.Printf("[%s/%s]: error fetching %s: %v", githubLogin, repoName, path, err)
		return false
	}

	return len(directoryContents) > 0 // check for at least one action
}

func containsFile(ctx context.Context, client *gh.Client, githubLogin string, repoName string, path string) bool {
	fileContents, _, resp, err := client.Repositories.GetContents(ctx, githubLogin, repoName, path, nil)
	if resp.StatusCode == http.StatusNotFound {
		return false
	}
	if err != nil {
		log.Printf("[%s/%s]: error fetching %s: %v", githubLogin, repoName, path, err)
		return false
	}

	return fileContents != nil
}

type dependabotFileStructure struct {
	Version int `yaml:"version"`
	Updates []struct {
		PackageEcosystem string `yaml:"package-ecosystem"`
	} `yaml:"updates"`
}

func readDependabotConfig(ctx context.Context, client *gh.Client, githubLogin string, repoName string) dependabotFileStructure {
	fileContents, _, _, err := client.Repositories.GetContents(ctx, githubLogin, repoName, ".github/dependabot.yml", nil)
	if err == nil {
		content, err := fileContents.GetContent()
		if err == nil {
			dependabotFile := &dependabotFileStructure{}
			if err := yaml.Unmarshal([]byte(content), &dependabotFile); err == nil {
				return *dependabotFile
			}
		}
	}
	return dependabotFileStructure{}
}
