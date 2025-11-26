#!/bin/bash
unset dirs files
FILTER_PATHS=(
    "/ethrpc/tracers"  
    "/ethrpc/debug"  
)
dirs=$(go list -f {{.Dir}} ./... | grep -v /vendor/)
for d in $dirs
do
    skip=0
    for filter in "${FILTER_PATHS[@]}"; do
        if [[ "$d" == *"$filter"* ]]; then
            skip=1
            break  
        fi
    done
    [[ $skip -eq 1 ]] && continue

    for f in $d/*.go
    do
    grep -q "Copyright (C) 20[1-2][0-9] The [o|O]ntology Authors" $f || files="${files} $f"
    done
done

ret=0
for f in $files
do
  echo "missing license:$f"
  ret=1
done

exit $ret
