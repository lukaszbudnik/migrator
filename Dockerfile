FROM golang:1.16.3-alpine3.13 as builder

LABEL maintainer="Łukasz Budnik lukasz.budnik@gmail.com"

ARG SOURCE_BRANCH
ARG SOURCE_COMMIT
ARG SOURCE_DATE

# build migrator
RUN mkdir -p /go/migrator
COPY . /go/migrator

RUN cd /go/migrator && \
  go build -ldflags "-X main.GitCommitDate=$SOURCE_DATE -X main.GitCommitSha=$SOURCE_COMMIT -X main.GitBranch=$SOURCE_BRANCH"

FROM alpine:3.13.4
COPY --from=builder /go/migrator/migrator /bin

VOLUME ["/data"]

# copy and register entrypoint script
COPY docker-entrypoint.sh /
ENTRYPOINT ["/docker-entrypoint.sh"]

EXPOSE 8080
