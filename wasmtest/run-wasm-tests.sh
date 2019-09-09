#!/bin/bash
set -e
set -x

# install build tools
if ! which rustup ; then
	curl https://sh.rustup.rs -sSf | sh -s -- -y --default-toolchain nightly 
	source $HOME/.cargo/env
fi
rustup target add wasm32-unknown-unknown
which ontio-wasm-build || cargo install --git=https://github.com/ontio/ontio-wasm-build

# build rust wasm contracts
mkdir -p testwasmdata
cd contracts-rust && bash travis.build.sh && cd ../

cd contracts-cplus && bash travis.build.bash && cd ../

# verify and optimize wasm contract
for wasm in testwasmdata/* ; do
	ontio-wasm-build $wasm $wasm
done

# start test framework
go run wasm-test.go
