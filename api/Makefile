# Useubash with error checking and pipeline failure detection
SHELL := /bin/bash
.SHELLFLAGS := -e -o pipefail -c
.PHONY: FORCE

RUN_TARGET ?= test

test: FORCE
	rm -rf app.db
	go test -v ./...
