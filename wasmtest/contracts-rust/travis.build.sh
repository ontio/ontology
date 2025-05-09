#!/bin/bash
set -e
set -x

# rm -rf js-vm && git clone --depth=1 https://github.com/laizy/ontio-js-vm js-vm
# cd js-vm
# RUSTFLAGS="-C link-arg=-zstack-size=32768" cargo build --lib --release --target=wasm32-unknown-unknown
# cd ..
# mv ./target/wasm32-unknown-unknown/release/boa.wasm ../testwasmdata/jsvm.wasm

for dir in $(ls)
do
	[[ -d $dir ]] && {
		cd $dir
		echo $dir
		RUSTFLAGS="-C target-cpu=mvp -C link-arg=-zstack-size=32768" cargo build -Zbuild-std=panic_abort,core,alloc --release --target wasm32-unknown-unknown
		cd ..
	}
done


cp ./target/wasm32-unknown-unknown/release/*.wasm ../testwasmdata

