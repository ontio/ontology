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
$CMD wallet -c --name $WALLET --password $PASSWD
if (( $? != 0 )); then
	echo "wallet creation failed"
	exit 1
fi

# list wallet
output=$($CMD wallet -l --name $WALLET --password $PASSWD)
if (( $? != 0 )); then
	echo "wallet listing failed"
	exit 1
fi
programhash=$(echo "$output" | grep "program hash" | awk -F : '{print $2}')

echo "$CMD asset --reg --name DNA --value 10000 --wallet $WALLET --password $PASSWD"
# asset registration
output=$($CMD asset --reg --name DNA --value 10000 --wallet $WALLET --password $PASSWD)
if (( $? != 0 )); then
	echo "asset registration failed"
	exit 1
fi
assetid=$(getHashFromOutput "$output")
echo "Asset ID: $assetid"

sleep 6

# asset issuance
output=$($CMD asset --issue --asset $assetid --to $programhash --value 9999 --wallet $WALLET --password $PASSWD)
if (( $? != 0 )); then
	echo "asset issuance failed"
	exit 1
fi
issue=$(getHashFromOutput "$output")
echo "Issue: $issue"

sleep 6

# asset transfer 
output=$(./nodectl asset --transfer --asset $assetid --to $programhash --value 1 --wallet $WALLET --password $PASSWD)
if (( $? != 0 )); then
	echo "asset transfer failed"
	exit 1
fi
transfer=$(getHashFromOutput "$output")
echo "Transfer: $transfer"

echo PASS

exit 0
