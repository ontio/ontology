#!/usr/bin/env bash
set -ex

VERSION=$(git describe --always --tags --long)

if [ $RUNNER_OS == 'Linux' ]; then
  echo "linux sys"
  env GO111MODULE=on make all
  cd ./wasmtest && bash ./gha-run-wasm-tests.sh && cd ../
  bash ./.gha.check-license.sh
  bash ./.gha.check-templog.sh
  bash ./.gha.gofmt.sh
  bash ./.gha.gotest.sh
elif [ $RUNNER_OS == 'osx' ]; then
  echo "osx sys"
  env GO111MODULE=on make all
else
  echo "win sys"
  env GO111MODULE=on CGO_ENABLED=1 go build  -ldflags "-X github.com/ontio/ontology/common/config.Version=${VERSION}" -o ontology-windows-amd64 main.go
  env GO111MODULE=on go build  -ldflags "-X github.com/ontio/ontology/common/config.Version=${VERSION}" -o sigsvr-windows-amd64 sigsvr.go
fi
