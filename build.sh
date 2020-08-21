#!/bin/sh

# Golang Environments
GO111MODULE=on
GOSUMDB=sum.golang.google.cn
GOPROXY=https://goproxy.cn,direct

# Build Version Information
COMMIT=$(git rev-parse HEAD 2>/dev/null)
VERSION=$(git describe --tags 2>/dev/null)
BUILD_DATE=$(date +"%s")

# Golang Build Flags
BUILD_FLAGS_DATE="-X github.com/xgfone/gover.BuildTime=$BUILD_DATE"
BUILD_FLAGS_COMMIT="-X github.com/xgfone/gover.Commit=$COMMIT"
BUILD_FLAGS_VERSION="-X github.com/xgfone/gover.Version=$VERSION"
BUILD_FLAGS_X="$BUILD_FLAGS_DATE $BUILD_FLAGS_COMMIT $BUILD_FLAGS_VERSION"

# Build App
go build -ldflags "-w $BUILD_FLAGS_X"
