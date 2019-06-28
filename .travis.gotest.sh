#!/bin/bash
# code from https://github.com/Seklfreak/Robyul2
unset dirs files
dirs=$(go list ./... | grep -v vendor/ | grep -v testsuite | grep -v ontology$)
set -x -e
for d in $dirs
do
        go test -v $d
done

testcases=('smartcontract' 'smartcontract/sys-contract' 'sync')
for d in ${testcases[*]}
do
        rm -rf testsuite/Chain
        go test -v github.com/ontio/ontology/testsuite/$d
done

