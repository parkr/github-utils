## github-change-default-branch

Iterate over a user/org's repos, and change the default branch of each repo to the name specified.

```text
Usage of github-change-default-branch:
  -login string
    	GitHub Login (user or org) whose repos to list (default: currently-authorized user)
  -new-name string
    	The new name to use for the default branch on given repos (default "main")
```

Example:

```console
$ github-change-default-branch -login=Microsoft -new-name=main
```

Installation:

```console
$ go get github.com/parkr/github-utils/cmd/github-change-default-branch
```
