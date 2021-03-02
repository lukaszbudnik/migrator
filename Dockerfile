FROM golang:1.16.0-alpine3.13 as builder

LABEL maintainer="Łukasz Budnik lukasz.budnik@gmail.com"

ARG SOURCE_BRANCH

# build migrator
RUN apk add git
RUN git clone https://github.com/lukaszbudnik/migrator.git
RUN cd /go/migrator && git checkout $SOURCE_BRANCH && \
  GIT_BRANCH=$(git branch | awk -v FS=' ' '/\*/{print $NF}' | sed 's|[()]||g') && \
  GIT_COMMIT_DATE=$(git log -n1 --date=iso-strict | grep 'Date:' | sed 's|Date:\s*||g') && \
  GIT_COMMIT_SHA=$(git rev-list -1 HEAD) && \
  go build -ldflags "-X main.GitCommitDate=$GIT_COMMIT_DATE -X main.GitCommitSha=$GIT_COMMIT_SHA -X main.GitBranch=$GIT_BRANCH"

FROM alpine:3.13.2
COPY --from=builder /go/migrator/migrator /bin

VOLUME ["/data"]

# copy and register entrypoint script
COPY docker-entrypoint.sh /
ENTRYPOINT ["/docker-entrypoint.sh"]

EXPOSE 8080
