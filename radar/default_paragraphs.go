package radar

import (
	"bufio"
	"context"
	"fmt"
	"regexp"
	"strings"
)

var (
	fromLineRegexp = regexp.MustCompile(`^From @(\S+)(\:|'s radar comments)`)
)

var graphQLQueryIssuesAndComments = `{
    repository(owner: "%s", name: "%s") {
        issues(labels: ["radar"], states: [OPEN], last: 1) {
          nodes {
            url
            body
            comments(first: 100) {
              nodes {
                author {
                  login
                }
                body
              }
            }
          }
        }
      }
}`

type issuesAndComments struct {
	Data struct {
		Repository struct {
			Issues struct {
				Nodes []struct {
					URL      string `json:"url"`
					Body     string `json:"body"`
					Comments struct {
						Nodes []struct {
							Author struct {
								Login string `json:"login"`
							} `json:"author"`
							Body string `json:"body"`
						} `json:"nodes"`
					} `json:"comments"`
				} `json:"nodes"`
			} `json:"issues"`
		} `json:"repository"`
	} `json:"data"`
}

func (r *Radar) AddDefaultParagraphs(ctx context.Context) error {
	r.AddParagraph(r.WhatsUpParagraph())
	r.AddParagraph(r.PreviouslyParagraph(ctx))
	r.AddParagraph("---")

	paragraphs, err := r.PreviousTasksParagraphs(ctx)
	if err != nil {
		return err
	}
	for _, previousTasks := range paragraphs {
		r.AddParagraph(previousTasks)
	}

	return nil
}

func (r *Radar) PreviouslyParagraph(ctx context.Context) string {
	if issue := r.GetPrevious(ctx); issue != nil {
		return fmt.Sprintf("*[Previously](%s).*", r.GetPrevious(ctx).GetHTMLURL())
	}
	return ""
}

func (r *Radar) WhatsUpParagraph() string {
	return fmt.Sprintf("What's up, %s? What happened? What's going to happen? What are you thinking?", r.mention)
}

func (r *Radar) PreviousTasksParagraphs(ctx context.Context) ([]string, error) {
	previousTasks := map[string][]string{}

	var data issuesAndComments
	req, err := r.github.NewRequest("POST", "graphql", struct {
		Query string `json:"query"`
	}{Query: fmt.Sprintf(graphQLQueryIssuesAndComments, r.repoOwner, r.repoName)})
	if err != nil {
		return []string{}, err
	}
	_, err = r.github.Do(ctx, req, &data)
	if err != nil {
		return []string{}, err
	}
	for _, issue := range data.Data.Repository.Issues.Nodes {
		r.parseBodyForTasks(issue.Body, previousTasks, "")
		for _, comments := range issue.Comments.Nodes {
			r.parseBodyForTasks(comments.Body, previousTasks, comments.Author.Login)
		}
	}

	var paragraphs []string
	for user, tasks := range previousTasks {
		paragraphs = append(paragraphs, fmt.Sprintf("%s\n\n%s",
			"From @"+user+":",
			strings.Join(tasks, "\n"),
		))
	}

	return paragraphs, nil
}

func (r *Radar) parseBodyForTasks(body string, dest map[string][]string, userOverride string) {
	scanner := bufio.NewScanner(strings.NewReader(body))
	uncheckedTasksUser := userOverride
	for scanner.Scan() {
		fmt.Println("line", scanner.Text())
		if results := fromLineRegexp.FindStringSubmatch(scanner.Text()); len(results) >= 1 {
			uncheckedTasksUser = results[1]
			continue
		}

		if strings.HasPrefix(scanner.Text(), "- [ ] ") {
			if _, ok := dest[uncheckedTasksUser]; !ok {
				dest[uncheckedTasksUser] = []string{}
			}

			dest[uncheckedTasksUser] = append(dest[uncheckedTasksUser], scanner.Text())
		}
	}
}
