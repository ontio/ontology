#!/bin/bash
set -e
set -x

# install build tools
# ensure clang

RUST_VERSION=nightly-2025-03-11

# ensure rust
curl https://sh.rustup.rs -sSf | sh -s -- -y --default-toolchain $RUST_VERSION
source $HOME/.cargo/env

rustup default $RUST_VERSION
rustup target add wasm32-unknown-unknown
which ontio-wasm-build || cargo install --git=https://github.com/ontio/ontio-wasm-build

# build rust wasm contracts
mkdir -p testwasmdata
cd contracts-rust && bash travis.build.sh && cd ../

# verify and optimize wasm contract
for wasm in testwasmdata/*.wasm ; do
  ontio-wasm-build $wasm $wasm
done

# start test framework
go build
./wasmtest
