#!/bin/bash
#
# Copyright (C) 2018 The ontology Authors
# This file is part of The ontology library.
#
# The ontology is free software: you can redistribute it and/or modify
# it under the terms of the GNU Lesser General Public License as published by
# the Free Software Foundation, either version 3 of the License, or
# (at your option) any later version.
#
# The ontology is distributed in the hope that it will be useful,
# but WITHOUT ANY WARRANTY; without even the implied warranty of
# MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
# GNU Lesser General Public License for more details.
#
# You should have received a copy of the GNU Lesser General Public License
# along with The ontology.  If not, see <http://www.gnu.org/licenses/>.
#

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


