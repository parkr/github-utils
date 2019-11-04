package contributions

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/url"
	"strings"
	"time"

	"github.com/google/go-github/v28/github"
	"github.com/parkr/github-utils/gh"
	"github.com/parkr/github-utils/search"
)

type contributionsTracker struct {
	login, startDate, owner string

	startDateAsTime time.Time

	github *gh.Client
}

func New(client *gh.Client, login, startDate, owner string) *contributionsTracker {
	startDateAsTime, err := time.Parse("2006-01-02", startDate)
	if err != nil {
		panic(err)
	}

	if owner != "" {
		owner = "@" + owner
	}
	return &contributionsTracker{
		login:           login,
		owner:           owner,
		startDate:       startDate,
		startDateAsTime: startDateAsTime,
		github:          client,
	}
}

func (c *contributionsTracker) Write(writer io.Writer) error {
	fmt.Fprintf(writer, "Contributions for %s since %s\n\n", c.login, c.startDate)

	if err := c.addPushedPRs(writer); err != nil {
		return err
	}

	if err := c.addShippedPRs(writer); err != nil {
		return err
	}

	if err := c.addTrackedIssues(writer); err != nil {
		return err
	}

	if err := c.addContributedIssues(writer); err != nil {
		return err
	}

	if err := c.addReviewedPRs(writer); err != nil {
		return err
	}

	return nil
}

func (c *contributionsTracker) String() (string, error) {
	var buf bytes.Buffer

	err := c.Write(&buf)

	return buf.String(), err
}

func (c *contributionsTracker) addPushedPRs(buf io.Writer) error {
	return c.addIssues(buf, "Pushed",
		fmt.Sprintf("created:>=%s %s author:%s type:pr state:open", c.startDate, c.owner, c.login),
		nil,
	)
}

func (c *contributionsTracker) addShippedPRs(buf io.Writer) error {
	return c.addIssues(buf, "Shipped",
		fmt.Sprintf("updated:>=%s %s author:%s type:pr state:closed", c.startDate, c.owner, c.login),
		func(issue github.Issue) bool {
			return c.gteStartTime(issue.ClosedAt)
		},
	)
}

func (c *contributionsTracker) addTrackedIssues(buf io.Writer) error {
	return c.addIssues(buf, "Tracked",
		fmt.Sprintf("created:>=%s %s author:%s type:issue", c.startDate, c.owner, c.login),
		nil,
	)
}

func (c *contributionsTracker) addContributedIssues(buf io.Writer) error {
	return c.addIssues(buf, "Contributed",
		fmt.Sprintf("updated:>=%s %s commenter:%s type:issue", c.startDate, c.owner, c.login),
		c.commentedInLastWeek,
	)
}

func (c *contributionsTracker) addReviewedPRs(buf io.Writer) error {
	return c.addIssues(buf, "Reviewed",
		fmt.Sprintf("updated:>=%s %s commenter:%s type:pr", c.startDate, c.owner, c.login),
		c.commentedInLastWeek,
	)
}

func (c *contributionsTracker) searchIssues(query string) ([]github.Issue, error) {
	var issues []github.Issue

	options := &github.SearchOptions{
		Sort:  "created",
		Order: "asc",
		ListOptions: github.ListOptions{
			Page:    0,
			PerPage: 100,
		},
	}
	for {
		result, resp, err := c.github.Search.Issues(context.Background(), query, options)
		if err != nil {
			return issues, err
		}

		issues = append(issues, result.Issues...)
		if resp.NextPage == 0 {
			break
		}
		options.Page = resp.NextPage
	}

	return issues, nil
}

func (c *contributionsTracker) addIssues(buf io.Writer, header, query string, filterFunc func(github.Issue) bool) error {
	unfilteredIssues, err := search.SearchIssues(c.github, query)
	if err != nil {
		return err
	}

	var issues []github.Issue
	if filterFunc != nil {
		for _, issue := range unfilteredIssues {
			if filterFunc(issue) {
				issues = append(issues, issue)
			}
		}
	} else {
		issues = unfilteredIssues
	}

	fmt.Fprintf(buf, "\n### %s (%d)\n", header, len(issues))
	for i := len(issues) - 1; i >= 0; i-- {

		fmt.Fprintf(buf, c.formattedIssue(issues[i])+"\n")
	}

	fmt.Fprintf(buf, "\n")

	return nil
}

func (c *contributionsTracker) formattedIssue(issue github.Issue) string {
	owner, name := repoNwo(issue)
	return fmt.Sprintf(" * [x] [%s/%s#%d](%s) %s",
		owner, name,
		issue.GetNumber(),
		issue.GetHTMLURL(),
		issue.GetTitle(),
	)
}

func (c *contributionsTracker) commentedInLastWeek(issue github.Issue) bool {
	// Issues the user authored will be in "Pushed" or "Shipped"
	if issue.User.GetLogin() == c.login {
		return false
	}

	// Issue was created in duration of interest, so all comments were too
	if c.gteStartTime(issue.CreatedAt) {
		return true
	}

	// Comment was posted by user in duration of interest
	if c.issueCommentsSinceStartDate(issue) {
		return true
	}

	// Pull request comment posted by user in duration of interest
	if c.prReviewCommentsSinceStartDate(issue) {
		return true
	}

	return false
}

func (c *contributionsTracker) issueCommentsSinceStartDate(issue github.Issue) bool {
	owner, name := repoNwo(issue)
	options := &github.IssueListCommentsOptions{
		Sort:      "created",
		Direction: "asc",
		Since:     c.startDateAsTime,
		ListOptions: github.ListOptions{
			Page:    0,
			PerPage: 100,
		},
	}
	for {
		comments, resp, err := c.github.Issues.ListComments(
			context.Background(),
			owner, name,
			issue.GetNumber(),
			options,
		)
		if err == nil {
			for _, comment := range comments {
				if comment.User.GetLogin() == c.login && c.gteStartTime(comment.CreatedAt) {
					return true
				}
			}
		}
		if resp.NextPage == 0 {
			break
		}
		options.Page = resp.NextPage
	}

	return false
}

func (c *contributionsTracker) prReviewCommentsSinceStartDate(issue github.Issue) bool {
	owner, name := repoNwo(issue)
	options := &github.PullRequestListCommentsOptions{
		Sort:      "created",
		Direction: "asc",
		Since:     c.startDateAsTime,
		ListOptions: github.ListOptions{
			Page:    0,
			PerPage: 100,
		},
	}
	for {
		comments, resp, err := c.github.PullRequests.ListComments(
			context.Background(),
			owner, name,
			issue.GetNumber(),
			options,
		)
		if err == nil {
			for _, comment := range comments {
				if comment.User.GetLogin() == c.login && c.gteStartTime(comment.CreatedAt) {
					return true
				}
			}
		}
		if resp.NextPage == 0 {
			break
		}
		options.Page = resp.NextPage
	}

	return false
}

func (c *contributionsTracker) gteStartTime(date *time.Time) bool {
	return date.After(c.startDateAsTime) || date.Equal(c.startDateAsTime)
}

func repoNwo(issue github.Issue) (string, string) {
	if issue.Repository != nil {
		return issue.Repository.GetOwner().GetLogin(),
			issue.Repository.GetName()
	}

	uri, err := url.Parse(issue.GetHTMLURL())
	if err != nil {
		panic(err)
	}
	uriPieces := strings.Split(uri.Path, "/")
	return uriPieces[1], uriPieces[2]
}
