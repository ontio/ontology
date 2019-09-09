#!/bin/bash
set -e
set -x

#rm -rf js-vm && git clone --depth=1 https://github.com/laizy/ontio-js-vm js-vm
cd js-vm
RUSTFLAGS="-C link-arg=-zstack-size=32768" cargo build --lib --release --target=wasm32-unknown-unknown
cd ..
mv ./target/wasm32-unknown-unknown/release/boa.wasm ../testwasmdata/jsvm.wasm

cd helloworld
RUSTFLAGS="-C link-arg=-zstack-size=32768" cargo build --release --target=wasm32-unknown-unknown
# use cargo test in root dir will cause compile error, should be a cargo bug
cargo test --features=mock 
cd ..

cp ./target/wasm32-unknown-unknown/release/*.wasm ../testwasmdata

