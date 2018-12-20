FROM golang:1.11.2-alpine3.8 as builder

MAINTAINER Łukasz Budnik lukasz.budnik@gmail.com

ARG SOURCE_BRANCH

# build migrator
RUN apk add git
RUN go get -d -v github.com/lukaszbudnik/migrator
RUN cd /go/src/github.com/lukaszbudnik/migrator && git checkout $SOURCE_BRANCH && ./setup.sh
RUN cd /go/src/github.com/lukaszbudnik/migrator && \
  GIT_BRANCH=$(git branch | awk -v FS=' ' '/\*/{print $NF}' | sed 's|[()]||g') && \
  GIT_COMMIT_SHA=$(git rev-list -1 HEAD) && \
  go build -ldflags "-X main.GitCommitSha=$GIT_COMMIT_SHA -X main.GitBranch=$GIT_BRANCH"

FROM alpine:3.8
COPY --from=builder /go/src/github.com/lukaszbudnik/migrator/migrator /bin

VOLUME ["/data"]

# copy and register entrypoint script
COPY docker-entrypoint.sh /
ENTRYPOINT ["/docker-entrypoint.sh"]

EXPOSE 8080
