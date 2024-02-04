SHELL := /bin/bash

ROOT := $(shell git rev-parse --show-toplevel)

VERSION := $(shell git describe --always --dirty=-dev)

.PHONY: build
build:
	@CGO_ENABLED=0 go build -ldflags="-X github.com/bmcustodio/kubectl-cilium/internal/version.Version=$(VERSION)" \
		-v -o "$(ROOT)/bin/kubectl-cilium" "$(ROOT)/cmd/kubectl-cilium/"

.PHONY: ci
ci: lint build

$(ROOT)/bin/golangci-lint:
	@curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s v1.55.2

.PHONY: lint
lint: $(ROOT)/bin/golangci-lint
	@$(ROOT)/bin/golangci-lint run --enable unused,errname,exhaustive,exportloopref,godot,gofumpt,goimports,makezero,paralleltest,prealloc,thelper,tparallel,unconvert,unparam,usestdlibvars,wastedassign --timeout 5m
