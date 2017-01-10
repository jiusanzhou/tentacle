.PHONY: default deps fmt server client release-server release-client all release-all clean contributors

GLIDE_GO_EXECUTABLE ?= go

BUILD_TAGS=debug

PWD=.

default: all

test:
	${GLIDE_GO_EXECUTABLE} test . ./gb ./path ./action ./tree ./util ./godep ./godep/strip ./gpm ./cfg ./dependency\
	 ./importer ./msg ./repo ./mirrors

bootstrap-dist:
	${GLIDE_GO_EXECUTABLE} get -u github.com/franciscocpg/gox
	cd ${GOPATH}/src/github.com/franciscocpg/gox && git checkout dc50315fc7992f4fa34a4ee4bb3d60052eeb038e
	cd ${GOPATH}/src/github.com/franciscocpg/gox && ${GLIDE_GO_EXECUTABLE} install

depss:
	go get -tags '$(BUILD_TAGS)' -d -v ${PWD}

deps:

fmt:
	${GLIDE_GO_EXECUTABLE} fmt github.com/jiusanzhou/tentacle

server: deps
	${GLIDE_GO_EXECUTABLE} build -tags '$(BUILD_TAGS)' ${PWD}/main/tentacled

client: deps
	${GLIDE_GO_EXECUTABLE} build -tags '$(BUILD_TAGS)' ${PWD}/main/tentacler

all: fmt server client

release-server: BUILD_TAGS=release
release-server: server

release-client: BUILD_TAGS=release
release-client: client

release-all: fmt release-server release-client

clean:
	go clean -i -r .

contributors:
	echo "Contributors to tentacles, both large and small:" > CONTRIBUTORS
	git log --raw | grep "^Author: " | sort | uniq | cut -d ' ' -f2- | sed 's/^/- /' | cut -d '<' -f1 >> CONTRIBUTORS