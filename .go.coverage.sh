#!/bin/bash

unset dirs files
dirs=$(go list ./... | grep -v vendor/ | grep -v ontology$)
set -x -e

pwd=`pwd`
tmpFile=$pwd/coverage.out.tmp
outFile=$pwd/coverage.out
rm -f $tmpFile
echo 'mode: count' > $outFile

for d in $dirs
do
	go test -v $d -coverprofile=$tmpFile
    if [ -f $tmpFile ]
    then
        cat $tmpFile | tail -n +2 >> $outFile 
        rm $tmpFile
    fi
done

go tool cover -html=$outFile -o coverage.html
go tool cover -func=$outFile -o coverage_func.txt

