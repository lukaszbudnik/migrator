FROM golang:1.18.3-alpine3.15 as builder

LABEL org.opencontainers.image.authors="Łukasz Budnik <lukasz.budnik@gmail.com>"

ARG GIT_REF
ARG GIT_SHA

# build migrator
RUN mkdir -p /go/migrator
COPY . /go/migrator

RUN cd /go/migrator && \
  go build -ldflags "-X main.GitSha=$GIT_SHA -X main.GitRef=$GIT_REF"

FROM alpine:3.16.0

LABEL org.opencontainers.image.authors="Łukasz Budnik <lukasz.budnik@gmail.com>"

COPY --from=builder /go/migrator/migrator /bin

VOLUME ["/data"]

# copy and register entrypoint script
COPY docker-entrypoint.sh /
ENTRYPOINT ["/docker-entrypoint.sh"]

EXPOSE 8080
