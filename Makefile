# Golang Env Variables
GOPROXY=https://goproxy.cn,direct
GOSUMDB=sum.golang.google.cn
GONOSUMDB=
GOPRIVATE=

NATIVE_GOOS:=$(shell go env GOHOSTOS)
NATIVE_GOARCH:=$(shell go env GOHOSTARCH)

APP:=$(shell go list ./cmd/...)

.PHONY: all build install download generate
all: build

install: download generate
	go install $(APP)

build: download generate
	go build -o bin/ $(APP)

generate:
	GOOS=$(NATIVE_GOOS) GOARCH=$(NATIVE_GOARCH) go generate ./cmd/...

download:
	go mod download
