FROM golang
WORKDIR /go/src/github.com/parkr/github-utils
COPY . /go/src/github.com/parkr/github-utils
RUN make

FROM scratch
COPY --from=0 /go/bin/* /bin/
