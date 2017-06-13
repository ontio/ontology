#!/bin/bash

. common.sh

WALLET1=wallet1.dat
WALLET2=wallet2.dat
PASSWD=testpasswd
DATA='hello world!'

# init
if [[ ! -f $CMD || ! -x $CMD ]]; then
	echo "$CMD is not exist"
	exit 1
fi

if [[ ! -f $CONFIG ]]; then
	echo "$CONFIG is not exist"
	exit 1
fi

# create wallet
$CMD wallet -c --name $WALLET1 --password $PASSWD
if (( $? != 0 )); then
	echo "wallet1 creation failed"
	exit 1
fi
$CMD wallet -c --name $WALLET2 --password $PASSWD
if (( $? != 0 )); then
	echo "wallet2 creation failed"
	exit 1
fi

# list wallet
output=$($CMD wallet -l --name $WALLET2 --password $PASSWD)
if (( $? != 0 )); then
	echo "wallet2 listing failed"
	exit 1
fi
pubkey=$(echo "$output" | grep "public key:" | awk '{print $3}')

#encrypt payload
output=$($CMD privacy --enc --wallet $WALLET1 --password $PASSWD --to $pubkey --data "$DATA")
if (( $? != 0 )); then
	echo "payload encryption failed"
	exit 1
fi
txhash=$(echo "$output" | grep "\"result\":" | awk '{print $2}' | sed 's/\"//g')

sleep 6

#decrypt payload
output=$($CMD privacy --dec --wallet $WALLET2 --password $PASSWD --txhash $txhash)
if (( $? != 0 )); then
	echo "payload decryption failed"
	exit 1
fi

plain=$(echo "$output" | grep "\"result\":" | awk '{sub(/[^ ]+ /, ""); print $0}' |sed 's/\"//g')

if [ "$plain" != "$DATA" ];then
	echo "data comparation failed"
	echo plain:"$plain"
	echo DATA:"$DATA"
	exit 1
else
	echo PASS
fi

exit 0


