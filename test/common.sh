#!/bin/bash

CMD=$(pwd)/nodectl
CONFIG=config.json
WALLET=wallet.dat
PASSWD=testpasswd

function getHashFromOutput()
{
	typeset output=$1
	echo "$output" | grep "result" | awk -F : '{print $2}' | tr -d \"
}
