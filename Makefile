# Useubash with error checking and pipeline failure detection
SHELL := /bin/bash
.SHELLFLAGS := -e -o pipefail -c
.PHONY: FORCE

RUN_TARGET ?= test

test: build FORCE
	rm -rf app.db
	go test -v --failfast ./...

build: FORCE
	go run .
	~/go/bin/goimports -w . 2>/dev/null || go install golang.org/x/tools/cmd/goimports@latest && ~/go/bin/goimports -w .
	gofmt -s -w .

