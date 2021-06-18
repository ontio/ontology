#!/bin/bash
# code from https://github.com/Seklfreak/Robyul2
# we redirect some golang pkg e.g golang.org/x/sys in go.mod
# so we need to change to empty repo to install goimports
which goimports || cd /tmp/; go get -v golang.org/x/tools/cmd/goimports; cd -

unset dirs files
dirs=$(go list -f {{.Dir}} ./... | grep -v /vendor/)

for d in $dirs
do
    for f in $d/*.go
    do
    files="${files} $f"
    done
done

diff <(goimports -d $files) <(echo -n)
