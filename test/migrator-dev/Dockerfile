FROM golang:1.17.0-alpine3.13 as builder

LABEL maintainer="Łukasz Budnik lukasz.budnik@gmail.com"

# use SOURCE_BRANCH to override at build time
ARG SOURCE_BRANCH=main

# git is required
RUN apk add git

# A - install migrator from local source code
RUN mkdir -p /go/migrator
COPY . /go/migrator

# B - install migrator from $SOURCE_BRANCH branch
# RUN git clone https://github.com/lukaszbudnik/migrator.git
# RUN cd /go/migrator && git checkout $SOURCE_BRANCH

RUN cd /go/migrator && \
  GIT_REF=$(git branch | awk -v FS=' ' '/\*/{print $NF}' | sed 's|[()]||g') && \
  GIT_SHA=$(git rev-list -1 HEAD) && \
  go build -ldflags "-X main.GitSha=$GIT_SHA -X main.GitRef=$GIT_REF"

FROM alpine:3.14.1
COPY --from=builder /go/migrator/migrator /bin

VOLUME ["/data"]

# copy and register entrypoint script
COPY docker-entrypoint.sh /
ENTRYPOINT ["/docker-entrypoint.sh"]

EXPOSE 8080
