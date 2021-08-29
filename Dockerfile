FROM golang:1.17.0-alpine3.13 as builder

LABEL maintainer="Łukasz Budnik lukasz.budnik@gmail.com"

ARG GIT_REF
ARG GIT_SHA

# build migrator
RUN mkdir -p /go/migrator
COPY . /go/migrator

RUN cd /go/migrator && \
  go build -ldflags "-X main.GitSha=$GIT_SHA -X main.GitRef=$GIT_REF"

FROM alpine:3.14.2
COPY --from=builder /go/migrator/migrator /bin

VOLUME ["/data"]

# copy and register entrypoint script
COPY docker-entrypoint.sh /
ENTRYPOINT ["/docker-entrypoint.sh"]

EXPOSE 8080
