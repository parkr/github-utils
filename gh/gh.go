package gh

import (
	"context"
	"fmt"
	"os"

	"github.com/bgentry/go-netrc/netrc"
	"github.com/google/go-github/github"
	"golang.org/x/oauth2"
)

var netrcFile = os.Getenv("HOME") + "/.netrc"
var netrcMachines = []string{"api.github.com", "github.com"}

type Client struct {
	*netrc.Machine
	*github.Client
	Context context.Context
}

func getLogin() (*netrc.Machine, error) {
	rc, err := netrc.ParseFile(netrcFile)
	if err != nil {
		return nil, err
	}

	for _, machineName := range netrcMachines {
		machine := rc.FindMachine(machineName)
		if !machine.IsDefault() {
			return machine, nil
		}
	}

	return nil, fmt.Errorf("no netrc config for any of: %s", netrcMachines)
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
	}, nil
}
