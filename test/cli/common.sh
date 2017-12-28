#!/bin/bash

CMD=$(pwd)/nodectl.exe
CONFIG=config.json
WALLET=walletT.dat
PASSWD=passwordtest

function getHashFromOutput()
{
	typeset output=$1
	echo "$output" | grep "result" | awk -F : '{print $2}' | tr -d \"
}
