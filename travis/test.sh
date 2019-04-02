#!/bin/bash
set -euC

function join_list {
    local IFS=","
    echo "$*"
}

case "$1" in
    "standard")
        go test -race $(go list ./... | grep -v /vendor/ | grep -v /pouchdb)
    ;;
    "gopherjs")
        unset KIVIK_TEST_DSN_COUCH16
        gopherjs test $(go list ./... | grep -v /vendor/ | grep -Ev 'kivik/(serve|auth|proxy)')
    ;;
    "linter")
        golangci-lint run ./...
    ;;
    "coverage")
        echo "" > coverage.txt

        for d in $(go list ./... | grep -v /vendor/ | grep -v /test/); do
            go test -race -coverprofile=profile.out -covermode=atomic $d
            if [ -f profile.out ]; then
                cat profile.out >> coverage.txt
                rm profile.out
            fi
        done

        bash <(curl -s https://codecov.io/bash)
esac
