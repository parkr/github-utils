package radar

import (
	"context"
	"log"
	"strings"
	"time"

	"github.com/parkr/github-utils/gh"

	"github.com/google/go-github/v50/github"
)

type RadarConfig struct {
	GitHub    *gh.Client
	Mention   string
	RepoOwner string
	RepoName  string
}

type Radar struct {
	github    *gh.Client
	mention   string
	repoOwner string
	repoName  string

	paragraphs []string
	previous   *github.Issue
}

func NewRadar(cfg *RadarConfig) *Radar {
	return &Radar{
		github:    cfg.GitHub,
		mention:   cfg.Mention,
		repoOwner: cfg.RepoOwner,
		repoName:  cfg.RepoName,
	}
}

func (r *Radar) GetPrevious(ctx context.Context) *github.Issue {
	if r.previous == nil {
		issues, _, err := r.github.Issues.ListByRepo(ctx, r.repoOwner, r.repoName, &github.IssueListByRepoOptions{
			State:       "open",
			Labels:      []string{"radar"},
			Sort:        "created",
			Direction:   "desc",
			ListOptions: github.ListOptions{PerPage: 1},
		})
		if err != nil {
			return nil
		}
		if len(issues) == 0 {
			return nil
		}
		r.previous = issues[0]
	}

	return r.previous
}

func (r *Radar) ClosePrevious(ctx context.Context) error {
	if r.previous == nil {
		return nil
	}

	_, _, err := r.github.Issues.Edit(ctx, r.repoOwner, r.repoName, r.previous.GetNumber(), &github.IssueRequest{
		State: github.String("closed"),
	})
	return err
}

func (r *Radar) Title() string {
	return "Radar: Week of " + time.Now().Format("2006-01-02")
}

func (r *Radar) AddParagraph(paragraph string) {
	if paragraph == "" {
		return
	}
	r.paragraphs = append(r.paragraphs, paragraph)
}

func (r *Radar) Body(ctx context.Context) string {
	if len(r.paragraphs) == 0 {
		if err := r.AddDefaultParagraphs(ctx); err != nil {
			log.Printf("error adding default paragraphs: %#v", err)
		}
	}

	return strings.Join(r.paragraphs, "\n\n")
}

func (r *Radar) Create(ctx context.Context) (*github.Issue, error) {
	// Make sure we populate the previous issue so we can close it later.
	r.GetPrevious(ctx)

	issue, _, err := r.github.Issues.Create(ctx, r.repoOwner, r.repoName, &github.IssueRequest{
		Title:  github.String(r.Title()),
		Body:   github.String(r.Body(ctx)),
		Labels: &[]string{"radar"},
	})
	return issue, err
}
