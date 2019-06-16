package langclassify

import (
	"log"

	"github.com/google/go-github/github"
	"github.com/parkr/github-utils/gh"
)

type Go struct {
	repo    *github.Repository
	version string
}

func (l *Go) GetName() string {
	return "Go"
}

func (l *Go) GetVersion(client *gh.Client) string {
	if l.version != "" {
		return l.version
	}
	// Check Gopkg.toml?
	return "unknown"
}

type Ruby struct {
	repo    *github.Repository
	version string
}

func (l *Ruby) GetName() string {
	return "Ruby"
}

func (l *Ruby) GetVersion(client *gh.Client) string {
	if l.version != "" {
		return l.version
	}
	// Check .ruby-version
	content, err := client.GetFileContents(l.repo.GetOwner().GetLogin(), l.repo.GetName(), ".ruby-version")
	if len(content) > 0 && err == nil {
		log.Printf("repo(%s): got a neat version from .ruby-version: %q", l.repo.GetFullName(), content)
		l.version = content
		return l.version
	}
	// Check Gemfile?
	return "unknown"
}

type JavaScript struct {
	version string
}

func (l *JavaScript) GetName() string {
	return "JavaScript"
}

func (l *JavaScript) GetVersion(client *gh.Client) string {
	if l.version != "" {
		return l.version
	}
	// Does package.json have this?
	return "unknown"
}

type Python struct {
	version string
}

func (l *Python) GetName() string {
	return "JavaScript"
}

func (l *Python) GetVersion(client *gh.Client) string {
	if l.version != "" {
		return l.version
	}
	return "unknown"
}
