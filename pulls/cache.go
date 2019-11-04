package pulls

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/laochailan/barely/maildir"
	"github.com/parkr/github-utils/gh"
)

// Once the PR is processed, a corresponding struct of this type is created
// to convey information about where the file went, whether it was successful, etc.
type OfflineStatusResponse struct {
	Success  bool
	Filename string
	Number   int
	Error    error
}

type Comments []Comment

func (c Comments) Len() int           { return len(c) }
func (c Comments) Less(i, j int) bool { return c[i].CreatedAt.Before(c[j].CreatedAt) }
func (c Comments) Swap(i, j int)      { c[i], c[j] = c[j], c[i] }

type Comment struct {
	To, Name, Username, Subject, Message string
	CreatedAt                            time.Time
}

func (c Comment) Email() string {
	return c.Username + "@users.noreply.github.com"
}

// Writes a single pull request data on the local disk. It pulls down:
//     1. A `.patch` file for the pull request.
//     1. A metadata file, preferably in Markdown or something human-readable.
//           - Username of author
//           - PR reviews & comments
//           - PR comment chain
func WritePullRequest(client *gh.Client, mailbox *maildir.Dir, repo string, pr PullRequest) OfflineStatusResponse {
	messages, err := WriteMaildirMessages(client, mailbox, repo, pr)
	if err != nil {
		return OfflineStatusResponse{Success: false, Error: err, Filename: "", Number: int(pr.Number)}
	}

	filename, _ := messages[0].Filename()
	return OfflineStatusResponse{Success: true, Error: nil, Filename: filename, Number: int(pr.Number)}
}

// Copies the content of the Patch URL for the PR down to a local file called <prNumber>.patch.
func GetPatch(client *gh.Client, pr PullRequest) (string, error) {
	patchURL := fmt.Sprintf("%s.patch", pr.Url)

	req, _ := http.NewRequest("GET", patchURL, nil)
	req.Header.Set("Authorization", "bearer "+client.Token)

	resp, err := client.HTTP.Do(req)
	if err != nil {
		return "", err
	}
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("bad response from %q: %s", patchURL, resp.Status)
	}

	patchContents, err := ioutil.ReadAll(resp.Body)
	defer resp.Body.Close()

	return string(patchContents), err
}

// Writes a file containing metadata for
func WriteMaildirMessages(client *gh.Client, mailbox *maildir.Dir, repo string, pr PullRequest) ([]*maildir.Message, error) {
	comments := Comments{}
	pieces := strings.Split(repo, "/")
	owner, repoName := pieces[0], pieces[1]

	var to string
	if hostname, err := os.Hostname(); err == nil {
		to += hostname
	} else {
		to += "localhost"
	}
	if currentUser := client.CurrentGitHubUser(); currentUser != nil {
		to = *currentUser.Login + "@" + to
	} else {
		to = "mbox@" + to
	}

	// Add the PR
	comments = append(comments, Comment{
		To:        to,
		Name:      userName(pr.Author),
		Username:  userLogin(pr.Author),
		Subject:   fmt.Sprintf("[#%d] %s", pr.Number, pr.Title),
		Message:   string(pr.Body),
		CreatedAt: pr.CreatedAt.Time,
	})

	// Add the patch
	patchContents, err := GetPatch(client, pr)
	if err != nil {
		log.Printf("unable to fetch patch for %s/%s#%d: %+v", owner, repoName, pr.Number, err)
	} else {
		comments = append(comments, Comment{
			To:        to,
			Name:      "",
			Username:  "",
			Subject:   fmt.Sprintf("RE: [#%d] %s", pr.Number, pr.Title),
			Message:   patchContents,
			CreatedAt: time.Now(),
		})
	}

	// Fetch the comments in the PR
	issueComments, err := GetPullRequestComments(client, owner, repoName, int(pr.Number))
	if err != nil {
		log.Printf("unable to fetch PR comments for %s/%s#%d: %+v", owner, repoName, pr.Number, err)
	}
	for _, comment := range issueComments {
		comments = append(comments, Comment{
			To:        to,
			Name:      userName(comment.Author),
			Username:  userLogin(comment.Author),
			Subject:   fmt.Sprintf("RE: [#%d] %s", pr.Number, pr.Title),
			Message:   string(comment.Body),
			CreatedAt: comment.CreatedAt.Time,
		})
	}

	// Fetch the line comments on the diff
	//prComments, err = GetPullRequestLineComments(client,pr)
	// for _, comment := range prComments {
	//     comments = append(comments, Comment{
	//         Name:      userName(comment.User),
	//         Username:  *comment.User.Login,
	//         Subject:   "Re: " + *pr.Title,
	//         Message:   *comment.Body,
	//         CreatedAt: *comment.CreatedAt,
	//     })
	// }

	// Sort the comments by when the comment was created
	sort.Stable(comments)

	messages := []*maildir.Message{}
	for _, comment := range comments {
		var buf bytes.Buffer
		_ = writeCommentAsMbox(comment, &buf)
		msg, err := mailbox.NewMessage(buf.Bytes())
		if err != nil {
			return messages, err
		}
		messages = append(messages, msg)
	}

	return messages, nil
}

// Write a comment to the buffer in mbox format.
func writeCommentAsMbox(comment Comment, buf io.Writer) error {
	mailtime := comment.CreatedAt.Format("Mon Jan 2 15:04:05 2006")
	mailtime2 := comment.CreatedAt.Format("Mon, 2 Jan 2006 15:04:05 -0700")
	fmt.Fprintf(buf, "From %s %s\n", comment.Email(), mailtime)
	fmt.Fprintf(buf, "Return-Path: <%s>\n", comment.Email())
	fmt.Fprintf(buf, "Delivered-To: %s\n", comment.To)
	fmt.Fprintf(buf, "Envelope-To: %s\n", comment.To)
	fmt.Fprintf(buf, "Delivery-Date: %s\n", mailtime2)
	fmt.Fprintf(buf, "From: %s\n", comment.Email())
	fmt.Fprintf(buf, "To: %s\n", comment.To)
	fmt.Fprintf(buf, "Subject: %s\n", comment.Subject)
	fmt.Fprintf(buf, "Date: %s\n", mailtime2)
	fmt.Fprintf(buf, "Status: RO\n")
	fmt.Fprintf(buf, "\n%s\n\n", comment.Message)
	return nil
}

func userLogin(author Author) string {
	if author.Bot.Login != "" {
		return string(author.Bot.Login)
	}
	return string(author.User.Login)
}

func userName(author Author) string {
	if author.User.Name != nil {
		return string(*author.User.Name)
	}
	return string(author.User.Login)
}
