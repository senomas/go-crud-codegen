# Useubash with error checking and pipeline failure detection
SHELL := /bin/bash
.SHELLFLAGS := -e -o pipefail -c
.PHONY: FORCE

RUN_TARGET ?= test

test: test-apigen build FORCE
	rm -rf app.db
	go test -v ./...

test-apigen: FORCE
	@cd ../apigen && go test ./...

build: FORCE
	@cd ../apigen && go run .
	@~/go/bin/goimports -w . 2>/dev/null || go install golang.org/x/tools/cmd/goimports@latest && ~/go/bin/goimports -w .
	@gofmt -s -w .


