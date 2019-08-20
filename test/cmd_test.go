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
		if _, err := os.Stat(filepath.Join(command, "/README.md")); err != nil {
			t.Errorf("expected a readme to exist for %s, but got error: %+v", command, err)
		}
	}
}
