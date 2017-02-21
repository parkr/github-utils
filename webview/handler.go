package webview

import (
	"fmt"
	"html/template"
	"net/http"
	"sort"
	"time"

	"github.com/google/go-github/github"
	"github.com/parkr/github-utils/gh"
	"github.com/parkr/github-utils/search"
)

type issueFetchFunc func(client *gh.Client, user string) ([]github.Issue, error)

type Page struct {
	Title  string
	Issues GitHubIssues
	Tmpl   *template.Template
}

type APIError struct {
	Err   error
	Title string
}

func (e APIError) Error() string {
	return fmt.Sprintf("%s: %+v", e.Title, e.Err)
}

type Handler struct {
	Users  []string
	Client *gh.Client
}

func (h Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.URL.Path {
	case "/":
		h.indexPage(w, r)
	case "/assigned":
		h.templatedPage(w, r, "Assigned", search.FindAllAssignedIssues)
	case "/mentions":
		h.templatedPage(w, r, "Mentioned", search.FindAllUnansweredMentions)
	case "/mine", "/authored", "/author":
		h.templatedPage(w, r, "Authored", search.FindAllCreatedIssues)
	default:
		http.Error(w, "path "+r.URL.Path+" not found", http.StatusNotFound)
	}
}

func (h Handler) boom(w http.ResponseWriter, err error) {
	errorTmpl.Execute(w, err)
	http.Error(w, err.Error(), http.StatusInternalServerError)
}

func (h Handler) indexPage(w http.ResponseWriter, r *http.Request) {
	name := "All Issues"

	pageChan := make(chan Page)
	errChan := make(chan *APIError)

	go h.newIssuesPage("Mentioned", search.FindAllUnansweredMentions, pageChan, errChan)
	go h.newIssuesPage("Assigned", search.FindAllAssignedIssues, pageChan, errChan)
	go h.newIssuesPage("Authored", search.FindAllCreatedIssues, pageChan, errChan)

	for i := 0; i < 3; i++ {
		select {
		case err := <-errChan:
			if err != nil {
				h.boom(w, err)
				return
			}
		case <-time.After(time.Second):
			// No error!
		}
	}

	h.renderPage(w, []Page{
		Page{Title: name, Tmpl: headerTmpl},
		<-pageChan,
		<-pageChan,
		<-pageChan,
		Page{Title: name, Tmpl: footerTmpl},
	})
	close(pageChan)
}

func (h Handler) templatedPage(w http.ResponseWriter, r *http.Request, name string, f issueFetchFunc) {
	issues, err := h.issuesForEveryone(f)
	if err != nil {
		h.boom(w, err)
		return
	}

	h.renderPage(w, []Page{
		Page{Title: name, Tmpl: headerTmpl},
		Page{Title: name, Tmpl: issuesTableTmpl, Issues: issues},
		Page{Title: name, Tmpl: footerTmpl},
	})
}

func (h Handler) renderPage(w http.ResponseWriter, pages []Page) {
	for _, page := range pages {
		if err := page.Tmpl.Execute(w, page); err != nil {
			h.boom(w, err)
			return
		}
	}
}

func (h Handler) newIssuesPage(name string, f issueFetchFunc, pageChan chan Page, errChan chan *APIError) {
	issues, err := h.issuesForEveryone(f)
	if err != nil {
		pageChan <- Page{Title: name, Tmpl: issuesTableTmpl, Issues: nil}
		errChan <- &APIError{Err: err, Title: name}
		return
	}

	pageChan <- Page{Title: name, Tmpl: issuesTableTmpl, Issues: issues}
	errChan <- nil
}

func (h Handler) issuesForEveryone(f issueFetchFunc) (GitHubIssues, error) {
	all := GitHubIssues{}

	for _, user := range h.Users {
		issues, err := f(h.Client, user)
		if err != nil {
			return nil, err
		}
		all = append(all, issues...)
	}

	sort.Stable(all)

	return all, nil
}
