package webview

import (
	"github.com/google/go-github/github"
)

type GitHubIssues []github.Issue

func (s GitHubIssues) Len() int {
	return len(s)
}

func (s GitHubIssues) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

// Issue i comes before j if i was created after j.
// Essentially CreatedAt descending, where the most recent is first.
func (s GitHubIssues) Less(i, j int) bool {
	return s[i].CreatedAt.Before(*s[j].CreatedAt)
}
