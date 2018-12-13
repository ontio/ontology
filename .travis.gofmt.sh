#!/bin/bash
# code from https://github.com/Seklfreak/Robyul2
unset dirs files
dirs=$(go list -f {{.Dir}} ./... | grep -v /vendor/)
    for d in $dirs
do
    for f in $d/*.go
    do
		files="${files} $f"
    done
done
diff <(gofmt -d $files) <(echo -n)
