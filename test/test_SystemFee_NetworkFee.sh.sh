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
# A word about this shell script:
#
# It must work everywhere, including on systems that lack
# a /bin/bash, map 'sh' to ksh, ksh97, bash, ash, or zsh,
# and potentially have either a posix shell or bourne
# shell living at /bin/sh.
#
# This shell is use to test all work with claim transactions and network fees.
# Claim transaction is used to get the system fee and system bounce.
# Network fee can speed up your transaction when transaction in waiting array.
# In this test, the Claim Result will print and the end, and it should be equal to the real Claim output.
# which can be verified by UTXO owned through rpc/restful/clent and so on.
# As a precondition of this test.
# 1.the max block transaction should be set to 5.
# 2.and the Reg/Issue transaction system fee should be set to 100.
# 3.use walletCreate.sh to create 11 wallet.
# The only shell it won't ever work on is cmd.exe.
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

# list wallet
output=$($CMD wallet -l --name $WALLET --password $PASSWD)
if (( $? != 0 )); then
    echo "wallet listing failed"
    exit 1
fi
programhash=$(echo "$output" | grep "program hash" | awk -F : '{print $2}')

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

##################### transfer gas to user account #######################
#
#
#
##########################################################################
height=$(getBlockHeight "$output")
echo "Transfer Gas to 11 account, Start height is $height"
output=$($CMD asset --transfer --asset $ONGAsset --to $programhash1 --value 10000000000 --wallet $WALLET  --netWorkFee 0 --password $PASSWD)
if (( $? != 0 )); then
    echo "asset transfer failed"
    exit 1
fi

transfer=$(getHashFromOutput "$output")
echo "Transfer ONG Hash: $transfer"
sleep 6

output=$($CMD asset --transfer --asset $ONGAsset --to $programhash2 --value 10000000000 --wallet $WALLET  --netWorkFee 0 --password $PASSWD)
if (( $? != 0 )); then
    echo "asset transfer failed"
    exit 1
fi

transfer=$(getHashFromOutput "$output")
echo "Transfer ONG Hash: $transfer"

sleep 6

output=$($CMD asset --transfer --asset $ONGAsset --to $programhash3 --value 10000000000 --wallet $WALLET  --netWorkFee 0 --password $PASSWD)
if (( $? != 0 )); then
    echo "asset transfer failed"
    exit 1
fi

transfer=$(getHashFromOutput "$output")
echo "Transfer ONG Hash: $transfer"
sleep 6

output=$($CMD asset --transfer --asset $ONGAsset --to $programhash4 --value 10000000000 --wallet $WALLET  --netWorkFee 0 --password $PASSWD)
if (( $? != 0 )); then
    echo "asset transfer failed"
    exit 1
fi

transfer=$(getHashFromOutput "$output")
echo "Transfer ONG Hash: $transfer"
sleep 6

output=$($CMD asset --transfer --asset $ONGAsset --to $programhash5 --value 10000000000 --wallet $WALLET  --netWorkFee 0 --password $PASSWD)
if (( $? != 0 )); then
    echo "asset transfer failed"
    exit 1
fi

transfer=$(getHashFromOutput "$output")
echo "Transfer ONG Hash: $transfer"
sleep 6

output=$($CMD asset --transfer --asset $ONGAsset --to $programhash6 --value 10000000000 --wallet $WALLET  --netWorkFee 0 --password $PASSWD)
if (( $? != 0 )); then
    echo "asset transfer failed"
    exit 1
fi

transfer=$(getHashFromOutput "$output")
echo "Transfer ONG Hash: $transfer"
sleep 6

output=$($CMD asset --transfer --asset $ONGAsset --to $programhash7 --value 10000000000 --wallet $WALLET  --netWorkFee 0 --password $PASSWD)
if (( $? != 0 )); then
    echo "asset transfer failed"
    exit 1
fi

transfer=$(getHashFromOutput "$output")
echo "Transfer ONG Hash: $transfer"
sleep 6

output=$($CMD asset --transfer --asset $ONGAsset --to $programhash8 --value 10000000000 --wallet $WALLET  --netWorkFee 0 --password $PASSWD)
if (( $? != 0 )); then
    echo "asset transfer failed"
    exit 1
fi

transfer=$(getHashFromOutput "$output")
echo "Transfer ONG Hash: $transfer"
sleep 6

output=$($CMD asset --transfer --asset $ONGAsset --to $programhash9 --value 10000000000 --wallet $WALLET  --netWorkFee 0 --password $PASSWD)
if (( $? != 0 )); then
    echo "asset transfer failed"
    exit 1
fi

transfer=$(getHashFromOutput "$output")
echo "Transfer ONG Hash: $transfer"
sleep 6

output=$($CMD asset --transfer --asset $ONGAsset --to $programhash10 --value 10000000000 --wallet $WALLET  --netWorkFee 0 --password $PASSWD)
if (( $? != 0 )); then
    echo "asset transfer failed"
    exit 1
fi

transfer=$(getHashFromOutput "$output")
echo "Transfer ONG Hash: $transfer"
sleep 6

output=$($CMD asset --transfer --asset $ONGAsset --to $programhash11 --value 10000000000 --wallet $WALLET  --netWorkFee 0 --password $PASSWD)
if (( $? != 0 )); then
    echo "asset transfer failed"
    exit 1
fi

transfer=$(getHashFromOutput "$output")
echo "Transfer ONG Hash: $transfer"
sleep 6
############################## asset registration #################################
#                                                                                 #
#                                                                                 #
#                                                                                 #
############################## asset registration #################################
height=$(getBlockHeight "$output")
echo "Registration start height is $height ."
output1=$($CMD asset --reg --name DNA1 --value 10000 --wallet $WALLET1 --netWorkFee 1100 --password $PASSWD)
if (( $? != 0 )); then
    echo "asset registration failed"
    exit 1
fi
assetid1=$(getHashFromOutput "$output1")

output2=$($CMD asset --reg --name DNA2 --value 10000 --wallet $WALLET2 --netWorkFee 1000 --password $PASSWD)
if (( $? != 0 )); then
    echo "asset registration failed"
    exit 1
fi
assetid2=$(getHashFromOutput "$output2")

output3=$($CMD asset --reg --name DNA3 --value 10000 --wallet $WALLET3 --netWorkFee 900 --password $PASSWD)
if (( $? != 0 )); then
    echo "asset registration failed"
    exit 1
fi
assetid3=$(getHashFromOutput "$output3")

output4=$($CMD asset --reg --name DNA4 --value 10000 --wallet $WALLET4 --netWorkFee 800 --password $PASSWD)
if (( $? != 0 )); then
    echo "asset registration failed"
    exit 1
fi
assetid4=$(getHashFromOutput "$output4")

output5=$($CMD asset --reg --name DNA5 --value 10000 --wallet $WALLET5 --netWorkFee 700 --password $PASSWD)
if (( $? != 0 )); then
    echo "asset registration failed"
    exit 1
fi
assetid5=$(getHashFromOutput "$output5")

output6=$($CMD asset --reg --name DNA6 --value 10000 --wallet $WALLET6 --netWorkFee 600 --password $PASSWD)
if (( $? != 0 )); then
    echo "asset registration failed"
    exit 1
fi
assetid6=$(getHashFromOutput "$output6")

output7=$($CMD asset --reg --name DNA7 --value 10000 --wallet $WALLET7 --netWorkFee 500 --password $PASSWD)
if (( $? != 0 )); then
    echo "asset registration failed"
    exit 1
fi
assetid7=$(getHashFromOutput "$output7")

output8=$($CMD asset --reg --name DNA8 --value 10000 --wallet $WALLET8 --netWorkFee 400 --password $PASSWD)
if (( $? != 0 )); then
    echo "asset registration failed"
    exit 1
fi
assetid8=$(getHashFromOutput "$output8")

output9=$($CMD asset --reg --name DNA9 --value 10000 --wallet $WALLET9 --netWorkFee 300 --password $PASSWD)
if (( $? != 0 )); then
    echo "asset registration failed"
    exit 1
fi
assetid9=$(getHashFromOutput "$output9")

output10=$($CMD asset --reg --name DNA10 --value 10000 --wallet $WALLET10 --netWorkFee 200 --password $PASSWD)
if (( $? != 0 )); then
    echo "asset registration failed"
    exit 1
fi
assetid10=$(getHashFromOutput "$output10")

output11=$($CMD asset --reg --name DNA11 --value 10000 --wallet $WALLET11 --netWorkFee 100 --password $PASSWD)
if (( $? != 0 )); then
    echo "asset registration failed"
    exit 1
fi
assetid11=$(getHashFromOutput "$output11")

echo "Test registration, Lots of registration Start..."
echo "Asset Reg Hash   : $assetid1"
echo "Asset Reg Hash   : $assetid2"
echo "Asset Reg Hash   : $assetid3"
echo "Asset Reg Hash   : $assetid4"
echo "Asset Reg Hash   : $assetid5"
echo "Asset Reg Hash   : $assetid6"
echo "Asset Reg Hash   : $assetid7"
echo "Asset Reg Hash   : $assetid8"
echo "Asset Reg Hash   : $assetid9"
echo "Asset Reg Hash   : $assetid10"
echo "Asset Reg Hash   : $assetid11"
sleep 18

############################## asset issue #################################
###                                                                      ###
############################## asset issue #################################
height=$(getBlockHeight "$output")
echo "Wait Process complete,Issue Start. Start height is $height ."
output=$($CMD asset --issue --asset $assetid1 --to $programhash1 --value 1 --wallet $WALLET1 --netWorkFee 0 --password $PASSWD)
if (( $? != 0 )); then
    echo "asset issuance failed"
    exit 1
fi
issue=$(getHashFromOutput "$output")
echo "Asset Issued Hash: $issue"
output=$($CMD asset --issue --asset $assetid2 --to $programhash2 --value 1 --wallet $WALLET2 --netWorkFee 0 --password $PASSWD)
if (( $? != 0 )); then
    echo "asset issuance failed"
    exit 1
fi
issue=$(getHashFromOutput "$output")
echo "Asset Issued Hash: $issue"
output=$($CMD asset --issue --asset $assetid3 --to $programhash3 --value 1 --wallet $WALLET3 --netWorkFee 0 --password $PASSWD)
if (( $? != 0 )); then
    echo "asset issuance failed"
    exit 1
fi
issue=$(getHashFromOutput "$output")
echo "Asset Issued Hash: $issue"
output=$($CMD asset --issue --asset $assetid4 --to $programhash4 --value 1 --wallet $WALLET4 --netWorkFee 0 --password $PASSWD)
if (( $? != 0 )); then
    echo "asset issuance failed"
    exit 1
fi
issue=$(getHashFromOutput "$output")
echo "Asset Issued Hash: $issue"
output=$($CMD asset --issue --asset $assetid5 --to $programhash5 --value 1 --wallet $WALLET5 --netWorkFee 0 --password $PASSWD)
if (( $? != 0 )); then
    echo "asset issuance failed"
    exit 1
fi
issue=$(getHashFromOutput "$output")
echo "Asset Issued Hash: $issue"
output=$($CMD asset --issue --asset $assetid6 --to $programhash6 --value 1 --wallet $WALLET6 --netWorkFee 0 --password $PASSWD)
if (( $? != 0 )); then
    echo "asset issuance failed"
    exit 1
fi
issue=$(getHashFromOutput "$output")
echo "Asset Issued Hash: $issue"
output=$($CMD asset --issue --asset $assetid7 --to $programhash7 --value 1 --wallet $WALLET7 --netWorkFee 0 --password $PASSWD)
if (( $? != 0 )); then
    echo "asset issuance failed"
    exit 1
fi
issue=$(getHashFromOutput "$output")
echo "Asset Issued Hash: $issue"
output=$($CMD asset --issue --asset $assetid8 --to $programhash8 --value 1 --wallet $WALLET8 --netWorkFee 0 --password $PASSWD)
if (( $? != 0 )); then
    echo "asset issuance failed"
    exit 1
fi
issue=$(getHashFromOutput "$output")
echo "Asset Issued Hash: $issue"
output=$($CMD asset --issue --asset $assetid9 --to $programhash9 --value 1 --wallet $WALLET9 --netWorkFee 0 --password $PASSWD)
if (( $? != 0 )); then
    echo "asset issuance failed"
    exit 1
fi
issue=$(getHashFromOutput "$output")
echo "Asset Issued Hash: $issue"
output=$($CMD asset --issue --asset $assetid10 --to $programhash10 --value 1 --wallet $WALLET10 --netWorkFee 0 --password $PASSWD)
if (( $? != 0 )); then
    echo "asset issuance failed"
    exit 1
fi
issue=$(getHashFromOutput "$output")
echo "Asset Issued Hash: $issue"
output=$($CMD asset --issue --asset $assetid11 --to $programhash11 --value 1 --wallet $WALLET11 --netWorkFee 0 --password $PASSWD)
if (( $? != 0 )); then
    echo "asset issuance failed"
    exit 1
fi
issue=$(getHashFromOutput "$output")
echo "Asset Issued Hash: $issue"
echo "Wait Process complete"
sleep 18
echo "Asset Issued Hash Completed."
########################### Issue complete   ################################
#
#
#
#############################################################################
#
#
#
########################### transfer Assets  ###############################
height=$(getBlockHeight "$output")
echo "Test netWorkFee, Lots of Transfer Start... Start height is $height . transaction hash is: "
output1=$($CMD asset --transfer --asset $assetid1 --to $programhash1 --value 1 --wallet $WALLET1  --netWorkFee 1100 --password $PASSWD)
if (( $? != 0 )); then
    echo "asset transfer failed"
    exit 1
fi
output2=$($CMD asset --transfer --asset $assetid2 --to $programhash1 --value 1 --wallet $WALLET2  --netWorkFee 1000 --password $PASSWD)
if (( $? != 0 )); then
    echo "asset transfer failed"
    exit 1
fi
output3=$($CMD asset --transfer --asset $assetid3 --to $programhash1 --value 1 --wallet $WALLET3  --netWorkFee 900 --password $PASSWD)
if (( $? != 0 )); then
    echo "asset transfer failed"
    exit 1
fi
output4=$($CMD asset --transfer --asset $assetid4 --to $programhash1 --value 1 --wallet $WALLET4  --netWorkFee 800 --password $PASSWD)
if (( $? != 0 )); then
    echo "asset transfer failed"
    exit 1
fi
output5=$($CMD asset --transfer --asset $assetid5 --to $programhash1 --value 1 --wallet $WALLET5  --netWorkFee 700 --password $PASSWD)
if (( $? != 0 )); then
    echo "asset transfer failed"
    exit 1
fi
output6=$($CMD asset --transfer --asset $assetid6 --to $programhash1 --value 1 --wallet $WALLET6  --netWorkFee 600 --password $PASSWD)
if (( $? != 0 )); then
    echo "asset transfer failed"
    exit 1
fi
output7=$($CMD asset --transfer --asset $assetid7 --to $programhash1 --value 1 --wallet $WALLET7  --netWorkFee 500 --password $PASSWD)
if (( $? != 0 )); then
    echo "asset transfer failed"
    exit 1
fi
output8=$($CMD asset --transfer --asset $assetid8 --to $programhash1 --value 1 --wallet $WALLET8  --netWorkFee 400 --password $PASSWD)
if (( $? != 0 )); then
    echo "asset transfer failed"
    exit 1
fi
output9=$($CMD asset --transfer --asset $assetid9 --to $programhash1 --value 1 --wallet $WALLET9  --netWorkFee 300 --password $PASSWD)
if (( $? != 0 )); then
    echo "asset transfer failed"
    exit 1
fi
output10=$($CMD asset --transfer --asset $assetid10 --to $programhash1 --value 1 --wallet $WALLET10  --netWorkFee 200 --password $PASSWD)
if (( $? != 0 )); then
    echo "asset transfer failed"
    exit 1
fi
output11=$($CMD asset --transfer --asset $assetid11 --to $programhash1 --value 1 --wallet $WALLET11  --netWorkFee 100 --password $PASSWD)
if (( $? != 0 )); then
    echo "asset transfer failed"
    exit 1
fi

## Print Transfer hash ##
transfer1=$(getHashFromOutput "$output1")
transfer2=$(getHashFromOutput "$output2")
transfer3=$(getHashFromOutput "$output3")
transfer4=$(getHashFromOutput "$output4")
transfer5=$(getHashFromOutput "$output5")
transfer6=$(getHashFromOutput "$output6")
transfer7=$(getHashFromOutput "$output7")
transfer8=$(getHashFromOutput "$output8")
transfer9=$(getHashFromOutput "$output9")
transfer10=$(getHashFromOutput "$output10")
transfer11=$(getHashFromOutput "$output11")
echo "Transfer1  Hash  : $transfer1"
echo "Transfer2  Hash  : $transfer2"
echo "Transfer3  Hash  : $transfer3"
echo "Transfer4  Hash  : $transfer4"
echo "Transfer5  Hash  : $transfer5"
echo "Transfer6  Hash  : $transfer6"
echo "Transfer7  Hash  : $transfer7"
echo "Transfer8  Hash  : $transfer8"
echo "Transfer9  Hash  : $transfer9"
echo "Transfer10 Hash  : $transfer10"
echo "Transfer11 Hash  : $transfer11"
sleep 24
########################### transfer Assets  Complete ###############################
#
#
#####################################################################################
height=$(getBlockHeight "$output")
echo "Asset Transfer completed. prepare for claim... Start height is $height"
# asset transfer
output=$($CMD asset --transfer --asset $ONTAsset --to $programhash --value 100000000000000000 --wallet $WALLET  --netWorkFee 0 --password $PASSWD)
if (( $? != 0 )); then
    echo "asset transfer failed"
    exit 1
fi

transfer=$(getHashFromOutput "$output")
echo "Transfer ONT Hash: $transfer"

sleep 6
# asset transfer
#
referTxID=33f50be9aa8caf36c03e84b8c23bc4dc9213d6de1a84f429c6a5611697418b95
output=$($CMD asset --claim --referTxID $referTxID --index 0 --wallet $WALLET --password $PASSWD)
if (( $? != 0 )); then
    echo "asset claim failed"
    exit 1
fi

transfer=$(getHashFromOutput "$output")
echo "Claim ONT Hash   : $transfer"

if (( $height <=50 )); then
    echo ">>>>>>>>>>Claim result should be $(expr $height  \* 80 + 2200)"
elif (( $height >50 && $height <=100)); then
    echo ">>>>>>>>>>Claim result should be $(expr 50 \* 80 + 2200 + $height \* 70  - 100 \* 70)"
else
    echo ">>>>>>>>>>Claim result should be $(expr  50 \* 80 + 50 \* 70 +  2200  + $height \* 60  - 100 \* 60)"
fi

echo PASS

exit 0