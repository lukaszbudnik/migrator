FROM golang:1.11.2-alpine3.8 as builder

MAINTAINER Łukasz Budnik lukasz.budnik@gmail.com

# install migrator
RUN apk add git
RUN go get github.com/lukaszbudnik/migrator

FROM alpine:3.8
COPY --from=builder /go/bin/migrator /bin

VOLUME ["/data"]

# copy and register entrypoint script
COPY docker-entrypoint.sh /
ENTRYPOINT ["/docker-entrypoint.sh"]

EXPOSE 8080
