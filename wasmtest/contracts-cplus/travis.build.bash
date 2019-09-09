#!/usr/bin/env bash
set -ev

oldir=$(pwd)
currentdir=$(dirname $0)
cd $currentdir

git clone https://github.com/ontio/ontology-wasm-cdt-cpp
compilerdir="./ontology-wasm-cdt-cpp/install/bin"

for f in $(ls *.cpp)
do
	$compilerdir/ont_cpp $f -lbase58 -lcrypto -lbuiltins -o  ${f%.cpp}.wasm
done

mv *.wasm ../testwasmdata/
rm *.wasm.str

cd $oldir
