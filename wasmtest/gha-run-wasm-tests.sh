#!/bin/bash
set -e
set -x

# install build tools
# ensure clang

wget releases.llvm.org/9.0.0/clang+llvm-9.0.0-x86_64-linux-gnu-ubuntu-18.04.tar.xz > /dev/null
tar xf clang+llvm-9.0.0-x86_64-linux-gnu-ubuntu-18.04.tar.xz > /dev/null 2>&1
export PATH="$(pwd)/clang+llvm-9.0.0-x86_64-linux-gnu-ubuntu-18.04/bin":$PATH

RUST_VERSION=nightly-2022-12-07

# ensure rust
curl https://sh.rustup.rs -sSf | sh -s -- -y --default-toolchain $RUST_VERSION
source $HOME/.cargo/env


rustup default $RUST_VERSION
rustup target add wasm32-unknown-unknown
which ontio-wasm-build || cargo install --git=https://github.com/ontio/ontio-wasm-build

# build rust wasm contracts
mkdir -p testwasmdata
cd contracts-rust && bash travis.build.sh && cd ../

cd contracts-cplus && bash travis.build.bash && cd ../

# verify and optimize wasm contract
for wasm in testwasmdata/*.wasm ; do
  ontio-wasm-build $wasm $wasm
done

# start test framework
go build
./wasmtest
