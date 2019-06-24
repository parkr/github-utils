workflow "Build on push" {
  on = "push"
  resolves = ["make"]
}

action "make" {
  uses = "parkr/actions/docker-make@master"
  args = ["make"]
}
