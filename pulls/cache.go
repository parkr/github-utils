package pulls

import (
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"sort"
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
	Name, Username, Subject, Message string
	CreatedAt                        time.Time
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
func WritePullRequest(client *gh.Client, outputDir string, pr *github.PullRequest) OfflineStatusResponse {
	patchFilename, err := WritePatchFile(client, outputDir, pr)
	if err != nil {
		return OfflineStatusResponse{Success: false, Error: err, Filename: patchFilename, Number: *pr.Number}
	}

	metadataFilename, err := WriteMetadataFile(client, outputDir, pr)
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
func WriteMetadataFile(client *gh.Client, outputDir string, pr *github.PullRequest) (string, error) {
	metadataFilename := filepath.Join(outputDir, fmt.Sprintf("%d.mbox", *pr.Number))

	f, err := os.OpenFile(metadataFilename, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0600)
	if err != nil {
		return metadataFilename, err
	}
	defer f.Close()

	comments := Comments{}

	// Add the PR
	comments = append(comments, Comment{
		Name:      userName(pr.User),
		Username:  *pr.User.Login,
		Subject:   *pr.Title,
		Message:   *pr.Body,
		CreatedAt: *pr.CreatedAt,
	})

	// Sort the comments by CreatedAt
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
	fmt.Fprintln(buf, "From "+comment.Email()+" "+mailtime)
	fmt.Fprintln(buf, "Return-path: <"+comment.Email()+">")
	fmt.Fprintln(buf, "Envelope-to: mbox@localhost")
	fmt.Fprintln(buf, "Delivery-date: "+mailtime2)
	fmt.Fprintln(buf, "From: "+comment.Email())
	fmt.Fprintln(buf, "To: mbox@localhost")
	fmt.Fprintln(buf, "Subject: "+comment.Subject)
	fmt.Fprintln(buf, "Date: "+mailtime2)
	fmt.Fprintln(buf, "\n"+comment.Message+"\n\n")
	return nil
}

func userName(user *github.User) string {
	if user.Name != nil {
		return *user.Name
	}
	return *user.Login
}
