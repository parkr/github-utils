package test

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
)

var allCommands []string

func getAllCommands() []string {
	if allCommands == nil {
		files, err := ioutil.ReadDir("../cmd")
		if err != nil {
			panic("couldn't list cmd: " + err.Error())
		}
		for _, f := range files {
			if f.IsDir() {
				allCommands = append(allCommands, f.Name())
			}
		}
	}

	return allCommands
}

func TestCmdHasReadme(t *testing.T) {
	for _, command := range getAllCommands() {
		f, err := os.Stat(filepath.Join("../cmd", command, "README.md"))
		if err != nil {
			t.Errorf("expected a readme to exist for %s, but got error: %+v", command, err)
		}
		if f != nil && f.Size() == 0 {
			t.Errorf("expected a readme to exist for %s, but it's empty", command)
		}
	}
}
