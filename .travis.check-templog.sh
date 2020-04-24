#!/bin/bash
unset dirs files
dirs=$(go list -f {{.Dir}} ./... | grep -v /vendor/)
for d in ${dirs}
do
    for f in ${d}/*.go
    do
		grep -q "log.Test" ${f} && files="${files} $f"
    done
done

ret=0
for f in ${files}
do
	echo "contains log.Test :$f"
	ret=1
done

exit ${ret}
