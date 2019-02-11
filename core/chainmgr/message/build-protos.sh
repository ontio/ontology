#!/usr/bin/env bash

# go get github.com/gogo/protobuf/protoc-gen-gogoslick
# add protoc-gen-gogoslick to your $PATH
protoc -I=. -I=$GOPATH/src --gogoslick_out=plugins=grpc:. protos.proto
