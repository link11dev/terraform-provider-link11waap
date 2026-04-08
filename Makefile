.DEFAULT_GOAL := help

SHELL := /bin/bash
CURRENT_DIR := $(shell pwd)

# golangci-lint config
golangci_lint_version=v2.9.0
vols=-v `pwd`:/app -w /app
run_lint=docker run --rm $(vols) golangci/golangci-lint:$(golangci_lint_version)

# terminal colors config
NO_COLOR=\033[0m
OK_COLOR=\033[32;01m

.PHONY: help build install test testacc lint fmt docs

## help: prints this help message
help:
	@echo "Usage:"
	@sed -n 's/^##//p' ${MAKEFILE_LIST} | column -t -s ':' |  sed -e 's/^/ /'

## build: builds the terraform provider binary
build:
	go build -o terraform-provider-link11waap

## install: builds and installs the terraform provider binary to the local LINUX terraform plugin directory
install: build
	mkdir -p ~/.terraform.d/plugins/registry.terraform.io/link11dev/link11waap/0.1.0/linux_amd64
	mv terraform-provider-link11waap ~/.terraform.d/plugins/registry.terraform.io/link11dev/link11waap/0.1.0/linux_amd64/

## test: runs all tests with verbose output
test:
	@printf "$(OK_COLOR)==> Running tests$(NO_COLOR)\n"
	@go test -v -count=1 -covermode=atomic -coverpkg=./... -coverprofile=coverage.txt ./...
	@go tool cover -func coverage.txt

## lint: run linters
lint:
	@printf "$(OK_COLOR)==> Running golang-ci-linter via Docker$(NO_COLOR)\n"
	@$(run_lint) golangci-lint run --max-issues-per-linter=0 --max-same-issues=0 --timeout=5m --verbose

## fmt: formats the code using gofmt
fmt:
	gofmt -s -w .

## docs: generates documentation for the terraform provider using tfplugindocs
docs:
	tfplugindocs generate --provider-name link11waap
