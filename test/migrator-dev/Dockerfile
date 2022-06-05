FROM golang:1.18.3-alpine3.15 as builder

LABEL org.opencontainers.image.authors="Łukasz Budnik <lukasz.budnik@gmail.com>"

# git is required
RUN apk add git

RUN mkdir -p /go/migrator
COPY . /go/migrator

RUN cd /go/migrator && go get -t -d ./...

RUN cd /go/migrator && \
  GIT_REF=$(git branch --show-current) && \
  GIT_SHA=$(git rev-parse HEAD) && \
  go build -o /bin/migrator -ldflags "-X main.GitSha=$GIT_SHA -X main.GitRef=$GIT_REF"

VOLUME ["/data"]

# copy and register entrypoint script
COPY docker-entrypoint.sh /
ENTRYPOINT ["/docker-entrypoint.sh"]

EXPOSE 8080
