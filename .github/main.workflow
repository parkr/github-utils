workflow "Build on push" {
  on = "push"
  resolves = ["docker build"]
}

action "docker build" {
  uses = "actions/docker/cli@master"
  args = ["build", "."]
}
