package contributions

import (
	"testing"

	"github.com/google/go-github/v62/github"
)

func TestRepoNwo(t *testing.T) {
	examples := []struct {
		input, expectedOwner, expectedName string
	}{
		{"https://github.com/github/github/pull/1234", "github", "github"},
		{"https://github.com/parkr/hello", "parkr", "hello"},
	}

	for _, example := range examples {
		actualOwner, actualName := repoNwo(github.Issue{HTMLURL: github.String(example.input)})
		if actualOwner != example.expectedOwner {
			t.Fatalf("input: %q, expected owner: %q, actual owner: %q",
				example.input, example.expectedOwner, actualOwner,
			)
		}
		if actualName != example.expectedName {
			t.Fatalf("input: %q, expected name: %q, actual name: %q",
				example.input, example.expectedName, actualName,
			)
		}
	}
}
