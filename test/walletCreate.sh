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
# #############################################################################################################
# this shell is use to create 11 account for default test.
#
# #############################################################################################################

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
$CMD wallet -c --name $WALLET1 --password $PASSWD
if (( $? != 0 )); then
    echo "wallet creation failed"
    exit 1
fi
# create wallet
$CMD wallet -c --name $WALLET2 --password $PASSWD
if (( $? != 0 )); then
    echo "wallet creation failed"
    exit 1
fi
# create wallet
$CMD wallet -c --name $WALLET3 --password $PASSWD
if (( $? != 0 )); then
    echo "wallet creation failed"
    exit 1
fi
# create wallet
$CMD wallet -c --name $WALLET4 --password $PASSWD
if (( $? != 0 )); then
    echo "wallet creation failed"
    exit 1
fi
# create wallet
$CMD wallet -c --name $WALLET5 --password $PASSWD
if (( $? != 0 )); then
    echo "wallet creation failed"
    exit 1
fi
# create wallet
$CMD wallet -c --name $WALLET6 --password $PASSWD
if (( $? != 0 )); then
    echo "wallet creation failed"
    exit 1
fi
# create wallet
$CMD wallet -c --name $WALLET7 --password $PASSWD
if (( $? != 0 )); then
    echo "wallet creation failed"
    exit 1
fi
# create wallet
$CMD wallet -c --name $WALLET8 --password $PASSWD
if (( $? != 0 )); then
    echo "wallet creation failed"
    exit 1
fi
# create wallet
$CMD wallet -c --name $WALLET9 --password $PASSWD
if (( $? != 0 )); then
    echo "wallet creation failed"
    exit 1
fi
# create wallet
$CMD wallet -c --name $WALLET10 --password $PASSWD
if (( $? != 0 )); then
    echo "wallet creation failed"
    exit 1
fi
# create wallet
$CMD wallet -c --name $WALLET11 --password $PASSWD
if (( $? != 0 )); then
    echo "wallet creation failed"
    exit 1
fi

# list wallet
output=$($CMD wallet -l --name $WALLET1 --password $PASSWD)
if (( $? != 0 )); then
    echo "wallet listing failed"
    exit 1
fi
programhash1=$(echo "$output" | grep "program hash" | awk -F : '{print $2}')

# list wallet
output=$($CMD wallet -l --name $WALLET2 --password $PASSWD)
if (( $? != 0 )); then
    echo "wallet listing failed"
    exit 1
fi
programhash2=$(echo "$output" | grep "program hash" | awk -F : '{print $2}')

# list wallet
output=$($CMD wallet -l --name $WALLET3 --password $PASSWD)
if (( $? != 0 )); then
    echo "wallet listing failed"
    exit 1
fi
programhash3=$(echo "$output" | grep "program hash" | awk -F : '{print $2}')

# list wallet
output=$($CMD wallet -l --name $WALLET4 --password $PASSWD)
if (( $? != 0 )); then
    echo "wallet listing failed"
    exit 1
fi
programhash4=$(echo "$output" | grep "program hash" | awk -F : '{print $2}')

# list wallet
output=$($CMD wallet -l --name $WALLET5 --password $PASSWD)
if (( $? != 0 )); then
    echo "wallet listing failed"
    exit 1
fi
programhash5=$(echo "$output" | grep "program hash" | awk -F : '{print $2}')

# list wallet
output=$($CMD wallet -l --name $WALLET6 --password $PASSWD)
if (( $? != 0 )); then
    echo "wallet listing failed"
    exit 1
fi
programhash6=$(echo "$output" | grep "program hash" | awk -F : '{print $2}')

# list wallet
output=$($CMD wallet -l --name $WALLET7 --password $PASSWD)
if (( $? != 0 )); then
    echo "wallet listing failed"
    exit 1
fi
programhash7=$(echo "$output" | grep "program hash" | awk -F : '{print $2}')

# list wallet
output=$($CMD wallet -l --name $WALLET8 --password $PASSWD)
if (( $? != 0 )); then
    echo "wallet listing failed"
    exit 1
fi
programhash8=$(echo "$output" | grep "program hash" | awk -F : '{print $2}')

# list wallet
output=$($CMD wallet -l --name $WALLET9 --password $PASSWD)
if (( $? != 0 )); then
    echo "wallet listing failed"
    exit 1
fi
programhash9=$(echo "$output" | grep "program hash" | awk -F : '{print $2}')

# list wallet
output=$($CMD wallet -l --name $WALLET10 --password $PASSWD)
if (( $? != 0 )); then
    echo "wallet listing failed"
    exit 1
fi
programhash10=$(echo "$output" | grep "program hash" | awk -F : '{print $2}')

# list wallet
output=$($CMD wallet -l --name $WALLET11 --password $PASSWD)
if (( $? != 0 )); then
    echo "wallet listing failed"
    exit 1
fi
programhash11=$(echo "$output" | grep "program hash" | awk -F : '{print $2}')

echo "Wallet Addr      :$programhash1"
echo "Wallet Addr      :$programhash2"
echo "Wallet Addr      :$programhash3"
echo "Wallet Addr      :$programhash4"
echo "Wallet Addr      :$programhash5"
echo "Wallet Addr      :$programhash6"
echo "Wallet Addr      :$programhash7"
echo "Wallet Addr      :$programhash8"
echo "Wallet Addr      :$programhash9"
echo "Wallet Addr      :$programhash10"
echo "Wallet Addr      :$programhash11"

echo PASS

exit 0
