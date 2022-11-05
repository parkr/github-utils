# github-dependabot-audit

Read a user/organizations repos (non-archived, non-forked), look for common
files related to Dependabot-updateable ecosystems, and indicate which
ecosystems are not covered by the `.github/dependabot.yml` file.

To check a single repo, pass the `-repo=name` parameter.

```shell
$ github-dependabot-audit -login=username [-repo=name]
```
