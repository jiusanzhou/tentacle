GO_EXECUTABLE ?= go
VERSION := $(shell git describe --always --long --dirty)

default: build-all

PWD = .

setup-build:
	go get github.com/mitchellh/gox
	go get github.com/Masterminds/glide

setup: setup-build

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

build-all: build-server build-client

.PHONY: build-all