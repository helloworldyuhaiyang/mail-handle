APP_VERSION:=beta
MODULE:=github.com/helloworldyuhaiyang/mail-handle
PROJECT:=mail-handle
IMAGE_NAME:=registry.cn-hangzhou.aliyuncs.com/helloworldyu/${PROJECT}:${APP_VERSION}
BIN_PATH:=../cmd/main.go

BUILD_TIME:=$(shell date "+%F %T")
COMMIT:=$(shell git rev-parse --verify HEAD)
LD:=-ldflags="-s -w -X '${MODULE}.AppName=${PROJECT}' -X '${MODULE}.Version=$(APP_VERSION)' -X '${MODULE}.Commit=$(COMMIT)' -X '${MODULE}.BuildDate=$(BUILD_TIME)'"

GOPATH:=$(shell go env GOPATH)

.PHONY: darwin
darwin:
	GOOS=darwin GOARCH=amd64 go build $(LD)  -v -o ./dist/${PROJECT}-darwin ${BIN_PATH}

.PHONY: linux
linux:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build $(LD) -v -o ./dist/${PROJECT}-linux ${BIN_PATH}

.PHONY: clean
clean:
	rm -rf ./dist
