package gh

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v2"

	"github.com/bgentry/go-netrc/netrc"
	"github.com/google/go-github/github"
	"golang.org/x/oauth2"
)

var netrcFile = filepath.Join(os.Getenv("HOME"), ".netrc")
var netrcMachines = []string{"api.github.com", "github.com"}
var hubConfigFile = filepath.Join(os.Getenv("HOME"), ".config", "hub")

type githubConnectionInfo struct {
	User       string `yaml:"user"`
	OauthToken string `yaml:"oauth_token"`
	Protocol   string `yaml:"protocol"`
}

type hubConfig struct {
	GitHub []githubConnectionInfo `yaml:"github.com"`
}

type Client struct {
	*netrc.Machine
	*github.Client
	Context context.Context

	currentlyAuthedGitHubUser *github.User
}

func loginFromNetrc(rc *netrc.Netrc) (*netrc.Machine, error) {
	for _, machineName := range netrcMachines {
		machine := rc.FindMachine(machineName)
		if machine != nil && !machine.IsDefault() {
			return machine, nil
		}
	}
	return nil, fmt.Errorf("no config for any of: %s", netrcMachines)
}

func loginFromNetrcFile() (*netrc.Machine, error) {
	rc, err := netrc.ParseFile(netrcFile)
	if err != nil {
		return nil, err
	}

	return loginFromNetrc(rc)
}

func loginFromHubConfigFile() (*netrc.Machine, error) {
	f, err := os.Open(hubConfigFile)
	if err != nil {
		return nil, err
	}

	cfg := &hubConfig{}
	yaml.NewDecoder(f).Decode(cfg)
	if len(cfg.GitHub) == 0 {
		return nil, fmt.Errorf("no config present in: %s", hubConfigFile)
	}

	rc := &netrc.Netrc{}
	for _, siteConf := range cfg.GitHub {
		rc.NewMachine("github.com", siteConf.User, siteConf.OauthToken, "")
	}

	return loginFromNetrc(rc)
}

func getLogin() (*netrc.Machine, error) {
	if machine, _ := loginFromNetrcFile(); machine != nil {
		return machine, nil
	}

	if machine, _ := loginFromHubConfigFile(); machine != nil {
		return machine, nil
	}

	return nil, fmt.Errorf("github login missing from %s and %s", netrcFile, hubConfigFile)
}

func NewDefaultClient() (*Client, error) {
	machine, err := getLogin()
	if err != nil {
		return nil, err
	}

	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: machine.Password},
	)
	tc := oauth2.NewClient(oauth2.NoContext, ts)

	return &Client{
		machine,
		github.NewClient(tc),
		context.Background(),
		nil,
	}, nil
}

func (c *Client) CurrentGitHubUser() *github.User {
	if c.currentlyAuthedGitHubUser == nil {
		currentlyAuthedUser, _, err := c.Users.Get(c.Context, "")
		if err != nil {
			log.Printf("couldn't fetch currently-auth'd user: %v", err)
			return nil
		}
		c.currentlyAuthedGitHubUser = currentlyAuthedUser
	}

	return c.currentlyAuthedGitHubUser
}

// GetFileContents fetches a single file and returns its contents.
func (c *Client) GetFileContents(repoOwner, repoName, filename string) (string, error) {
	// Check Dockerfile
	contentInfo, _, resp, err := c.Repositories.GetContents(c.Context, repoOwner, repoName, filename, nil)
	if err != nil {
		// Allow 404's to pass through.
		if resp.StatusCode == 404 {
			return "", nil
		}
		return "", err
	}

	return contentInfo.GetContent()
}
