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

