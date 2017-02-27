package pulls

import (
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/google/go-github/github"
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
	return c.Username + "@github"
}

// Writes a single pull request data on the local disk. It pulls down:
//     1. A `.patch` file for the pull request.
//     1. A metadata file, preferably in Markdown or something human-readable.
//           - Username of author
//           - PR reviews & comments
//           - PR comment chain
func WritePullRequest(client *gh.Client, outputDir string, repo string, pr *github.PullRequest) OfflineStatusResponse {
	patchFilename, err := WritePatchFile(client, outputDir, pr)
	if err != nil {
		return OfflineStatusResponse{Success: false, Error: err, Filename: patchFilename, Number: *pr.Number}
	}

	metadataFilename, err := WriteMetadataFile(client, outputDir, repo, pr)
	if err != nil {
		return OfflineStatusResponse{Success: false, Error: err, Filename: metadataFilename, Number: *pr.Number}
	}

	return OfflineStatusResponse{Success: true, Error: nil, Filename: patchFilename, Number: *pr.Number}
}

// Copies the content of the Patch URL for the PR down to a local file called <prNumber>.patch.
func WritePatchFile(client *gh.Client, outputDir string, pr *github.PullRequest) (string, error) {
	patchFilename := filepath.Join(outputDir, fmt.Sprintf("%d.patch", *pr.Number))

	resp, err := http.Get(*pr.PatchURL)
	if err != nil {
		return patchFilename, err
	}

	patchContents, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return patchFilename, err
	}
	defer resp.Body.Close()

	return patchFilename, ioutil.WriteFile(patchFilename, patchContents, 0600)
}

// Writes a file containing metadata for
func WriteMetadataFile(client *gh.Client, outputDir string, repo string, pr *github.PullRequest) (string, error) {
	metadataFilename := filepath.Join(outputDir, fmt.Sprintf("%d.mbox", *pr.Number))

	f, err := os.OpenFile(metadataFilename, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0600)
	if err != nil {
		return metadataFilename, err
	}
	defer f.Close()

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
		Name:      userName(pr.User),
		Username:  *pr.User.Login,
		Subject:   *pr.Title,
		Message:   *pr.Body,
		CreatedAt: *pr.CreatedAt,
	})

	// Fetch the comments in the PR
	issueComments, err := GetPullRequestComments(client, owner, repoName, *pr.Number)
	for _, comment := range issueComments {
		comments = append(comments, Comment{
			To:        to,
			Name:      userName(comment.User),
			Username:  *comment.User.Login,
			Subject:   "Re: " + *pr.Title,
			Message:   *comment.Body,
			CreatedAt: *comment.CreatedAt,
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

	for _, comment := range comments {
		if err := writeCommentAsMbox(comment, f); err != nil {
			return metadataFilename, err
		}
	}

	return metadataFilename, nil
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

func userName(user *github.User) string {
	if user.Name != nil {
		return *user.Name
	}
	return *user.Login
}

func unifiedCommentBody(comment *github.PullRequestComment) string {
	if comment.DiffHunk == nil {
		return *comment.Body
	}
	// TODO: fix this so the diff comment makes sense
	return *comment.DiffHunk + "\n\n" + *comment.Body
}
