workflow "Build on push" {
  on = "push"
  resolves = "docker build"
}

action "docker build" {
  uses = "docker"
  args = ["build", "."]
}
