SHELL := /bin/bash
.SHELLFLAGS := -e -o pipefail -c
.PHONY: FORCE
.ONESHELL:

$(shell test -f .local.env || touch .local.env)

include .env
-include .local.env

include ./docker.mk

build: FORCE
	@$(call docker-build,build-util,BUILD_UTIL,codegen-build-util)
	@$(call docker-build,.env .local.env Dockerfile entrypoint.sh base postgresql db2 *.go,CRUD_GEN,crudgen,Dockerfile,.)

test: FORCE
	rm -rf app.db
	go run -C ../crud-codegen/ . $(shell pwd) sqlite hanoman.co.id/crudgen
	go test -v --failfast ./...

