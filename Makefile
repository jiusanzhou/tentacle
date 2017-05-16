GO_EXECUTABLE ?= go
VERSION := $(shell git describe --always --long --dirty)

default: build

PWD = .

GOOS=$(shell go env GOOS)
GOARCH=$(shell go env GOARCH)

setup-build:
	go get github.com/mitchellh/gox

setup: setup-build

build-server-one: setup
	gox -verbose \
	-ldflags "-X main.version=${VERSION}" \
	-os="${GOOS}" \
	-arch="${GOARCH}" \
	-output="bin/{{.Dir}}" ${PWD}/main/tentacled

build-client-one: setup
	gox -verbose \
	-ldflags "-X main.version=${VERSION}" \
	-os="${GOOS}" \
	-arch="${GOARCH}" \
	-output="bin/{{.Dir}}" ${PWD}/main/tentacler

build: build-server-one build-client-one

build-server: setup
	gox -verbose \
	-ldflags "-X main.version=${VERSION}" \
	-os="linux darwin windows" \
	-arch="amd64 386" \
	-output="dist/tentacled-{{.OS}}-{{.Arch}}/{{.Dir}}" ${PWD}/main/tentacled

build-client: setup
	gox -verbose \
	-ldflags "-X main.version=${VERSION}" \
	-os="linux darwin windows" \
	-arch="amd64 386" \
	-output="dist/tentacler-{{.OS}}-{{.Arch}}/{{.Dir}}" ${PWD}/main/tentacler

all: build-server build-client

.PHONY: all build setup