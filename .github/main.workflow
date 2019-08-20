workflow "Build on push" {
  on = "push"
  resolves = ["docker build"]
}

action "docker build" {
  uses = "actions/docker/cli@master"
  args = ["build", "."]
}

workflow "Test on push" {
  on = "push"
  resolves = ["go test"]
}

action "go test" {
  uses = "actions/docker/cli-multi@master"
  args = [
    "build -f Dockerfile.cibuild -t github-utils-test .",
    "run github-utils-test make test"
  ]
}
