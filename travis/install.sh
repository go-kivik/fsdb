#!/bin/bash
set -euC
set -o xtrace

mkdir -p $(go env GOPATH)/bin
curl https://raw.githubusercontent.com/golang/dep/master/install.sh | sh
dep ensure

function generate {
    go get -u github.com/jteeuwen/go-bindata/...
    go generate $(go list ./... | grep -v /vendor/)
}

case "$1" in
    "standard")
        generate
    ;;
    "linter")
        curl -sfL https://install.goreleaser.com/github.com/golangci/golangci-lint.sh | sh -s -- -b $(go env GOPATH)/bin v1.17.1
    ;;
    "coverage")
        generate
    ;;
esac
