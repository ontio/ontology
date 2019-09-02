#!/bin/bash
set -e
set -x

RUSTFLAGS="-C link-arg=-zstack-size=32768" cargo build --release --target=wasm32-unknown-unknown

cd helloworld
# use cargo test in root dir will cause compile error, should be a cargo bug
cargo test --features=mock 
cd ..

cp ./target/wasm32-unknown-unknown/release/*.wasm ../testwasmdata

