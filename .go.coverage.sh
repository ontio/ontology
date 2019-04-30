#!/bin/bash

unset dirs files
dirs=$(go list ./... | grep -v vendor/ | grep -v ontology$)

mkdir coverage
tmpFile=./coverage/coverage.out.tmp
outFile=./coverage/coverage.out
testMissedFile=./coverage/coverage_missed.out
rm -f $tmpFile $testMissedFile
echo 'mode: count' > $outFile

for d in $dirs
do
	go test -v $d -coverprofile=$tmpFile
    if [ -f $tmpFile ]
    then
        cat $tmpFile | tail -n +2 >> $outFile 
        rm $tmpFile
    else
        echo -e "[no test files] \t" $d >> $testMissedFile
    fi
done

go tool cover -html=$outFile -o ./coverage/coverage.out.html
go tool cover -func=$outFile -o ./coverage/coverage_func.out.txt

