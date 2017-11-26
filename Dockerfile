FROM golang:1.9.2-alpine

RUN set -ex \
  && apk add --no-cache \
    curl \
    gcc \
    git \
    make \
    musl-dev

RUN set -ex \
  && go get -u github.com/golang/dep/cmd/dep \
  && go get -u github.com/goreleaser/goreleaser

WORKDIR /go/src/baku