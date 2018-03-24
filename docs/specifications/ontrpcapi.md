# ONT Rpc Api

* [Introduction](#Introduction)
* [Rpc api list](#Rpc api list)
* [Error code](#Error code)

## Introduction

Request parameter description:

| Field | Type | Description |
| :--- | :--- | :--- |
| jsonrpc | string | jsonrpc version |
| method | string | method name |
| params | string | method required parameters |
| id | int | any value |

Response parameter description:

| Field | Type | Description |
| :--- | :--- | :--- |
| desc| string | resopnse description |
| error | int64 | error code |
| jsonrpc | string | jsonrpc version |
| id | int | any value |
| result | object/string/bool | program execution result |

Note: The type of result varies with the request.

Block field description

| Field | Type | Description |
| :--- | :--- | :--- |
| Version | int | version number |
| PrevBlock | UInt256 | The hash of the previous block |
| TransactionsRoot | UInt256 | The root of the Merkle tree for all transactions in this block |
| BlockRoot | UInt256 | blockroot |
| Timestamp | int | block timestamp,uinix timestamp |
| Height | int | block height |
| NextBookKeeper | UInt160 | Accounting contract hash value for the next block |
| BookKeepers |  ||
| SigData |||
| Hash | Program | Script to verify the block |
| Transactions | Transaction[] | List of transactions in this block |

Transaction field description

| Field | Type | Description |
| :--- | :--- | :--- |
| Version| int | version number |
| TxType | TransactionType | transaction type |
| Payload | Payload | payload |
| Nounce | int | random number |
| Attributes | Transactions |  |
| Fee | Fee[] | transaction fees  |
| NetworkFee | long | neitwork fees |
| Sigs | Sign[] | signature array |
| Hash | string | transaction hash |

## Rpc api list

| Method | Parameters | Description | Note |
| :--- | :--- | :--- | :--- |
| getbestblockhash |  | get the hash of the highest height block in the main chain |  |
| getblock | <height> or <blockhash> [verbose] | get block by block height or block height | verbose can be 0 or 1,response is different |
| getblockcount |  | get the number of blocks |  |
| getblockhash | <index> | get block hash by index |  |
| getconnectioncount |  | get the current number of connections for the node |  |
| getrawmempool |  | Get a list of unconfirmed transactions in memory |  |
| getrawtransaction | <transactionhash> | Returns the corresponding transaction information based on the specified hash value. |  |
| sendrawtransaction | <hex> | Broadcast transaction. | Serialized signed transactions constructed in the program into hexadecimal strings |
| getstorage | <script_hash> | Returns the stored value according to the contract script hashes and stored key. |  |
| getversion |  | Get the version information of the query node |  |
| getblocksysfee |  | According to the specified index, return the system fee before the block. |  |
| getcontractstate | <script_hash> | According to the contract script hash, query the contract information. |  |
| getmempooltxstate | <tx_hash> | Query the transaction status in the memory pool. |  |
| getsmartcodeevent |  | Get smartcode event |  |
| getblockheightbytxhash | <tx_hash> | return balance of base58 account address. |  |
| getbalance | <address> | return balance of base58 account address. |  |


### 1. getbestblockhash

Get the hash of the highest height block in the main chain.

#### Example

Request:

```
{
  "jsonrpc": "2.0",
  "method": "getbestblockhash",
  "params": [],
  "id": 1
}
```

Response:

```
{
  "desc":"SUCCESS",
  "error":0,
  "jsonrpc": "2.0",
  "id": 1,
  "result": "773dd2dae4a9c9275290f89b56e67d7363ea4826dfd4fc13cc01cf73a44b0d0e"
}
```

Response instruction:

Result: The hash of the highest block height in the main chain.

### 2. getblock

Get the block information by block hash or height.

#### Parameter instruction

| Parameter | Type | Optional/required | Description |
| :--- | :--- | :--- | :--- |
| Hash/height | String/long | Required | Block hash/height |
| Verbose | int | Optional | Optional parameter, the default value of verbose is 0. When verbose is 0, it returns the block serialized information, which is represented by a hexadecimal string. To get detailed information from it, you need to call the SDK to deserialize. When verbose is 1, the detailed information of the corresponding block is returned, which is represented by a JSON format string. |

#### Example

Request:

```
{
  "jsonrpc": "2.0",
  "method": "getblock",
  "params": ["773dd2dae4a9c9275290f89b56e67d7363ea4826dfd4fc13cc01cf73a44b0d0e"],
  "id": 1
}
or
{
  "jsonrpc": "2.0",
  "method": "getblock",
  "params": [100],
  "id": 1
}
```

Response when verbose is nil:

```
{
  "desc":"SUCCESS",
  "error":0,
  "jsonrpc": "2.0",
  "id": 1,
  "result": "00000000ccc7612928aab25db55ab31c35c64929ce4d89f9a16d0753fddf9da63d0c339b77be0e825f3180b4d706045e42a101f5becea5d59a7d6aac58cdff0c0bd0b6a949c6405eae477bb053406c0a4f56a830289798e2d70dc77e0a1d927fa9fb93c47625f316f1bb594150e0f4c3b4c4c6394e0444f876c766b0130527ac46c766b0130c3648c00616c766b51c3c0519c009c6c766b0131527ac46c766b0131c3641000616c766b52c30052c461625400616c766b51c300c36c766b0132527ac46c766b0132c36165b3206c..."
}
```

Response when verbose = 1:

```
{
    "desc": "SUCCESS",
    "error": 0,
    "id": 1,
    "jsonpc": "2.0",
    "result": {
        "Hash": "71d614dd1594c0e4cc615e12678f6357c3686387519731b8231699301960c39d",
        "Header": {
            "Version": 0,
            "PrevBlockHash": "a12bc630d58af7c7ef94cc5339567f9042c95a9b5722d58dc3b87dd30bc6154e",
            "TransactionsRoot": "291026c1df3183e96839d0f0caf8c2640919de6b79114a026d35f73b4b6b3b48",
            "BlockRoot": "8f592ec921a950553be88a2f575b5c52f1ff0b9d5f263fbae86a5d1abd557dba",
            "Timestamp": 1521703551,
            "Height": 4,
            "ConsensusData": 17166119660593720000,
            "NextBookKeeper": "027c557d2e735b9a369d20dd099bfd42db5cdb74",
            "BookKeepers": [
                {
                    "X": "11045594958442581564679839478917319740817938700262919124154204990772552987783",
                    "Y": "28445199876541353997545685344458930058882115795876754515124389392470701852812"
                }
            ],
            "SigData": [
                "e8f2333b43ead2af0890edb8f104b5bba0b57a7192c30919ca8cc50dcc54890483f026c7733544d877fcf4ed76bc00dac90a000a4067347c99e593067e32bf19"
            ],
            "Hash": "71d614dd1594c0e4cc615e12678f6357c3686387519731b8231699301960c39d"
        },
        "Transactions": [
            {
                "Version": 0,
                "Nonce": 0,
                "TxType": 0,
                "Payload": {
                    "Nonce": 1521703551136164000,
                    "Issuer": {
                        "X": "",
                        "Y": ""
                    }
                },
                "Attributes": [ ],
                "Fee": null,
                "NetworkFee": 0,
                "Sigs": [
                    {
                        "PubKeys": [
                            {
                                "X": "11045594958442581564679839478917319740817938700262919124154204990772552987783",
                                "Y": "28445199876541353997545685344458930058882115795876754515124389392470701852812"
                            }
                        ],
                        "M": 1,
                        "SigData": [
                            "2531945f93ee57b651ef94bd91f10bc75dc539ff1f0afba32adfa59778b8f67ce4865094100e9697d4523e8a533791d2e0aa893d0a941a997add1f2ed5dfa338"
                        ]
                    }
                ],
                "Hash": "291026c1df3183e96839d0f0caf8c2640919de6b79114a026d35f73b4b6b3b48"
            }
        ]
    }
}
```

### 3. getblockcount

Get the number of blocks in the main chain.

#### Example

Request:

```
{
  "jsonrpc": "2.0",
  "method": "getblockcount",
  "params": [],
  "id": 1
}
```

Response:

```
{
  "desc":"SUCCESS",
  "error":0,
  "jsonrpc": "2.0",
  "id": 1,
  "result": 2519
}
```

Response instruction:

Result: the height of the main chain.

#### 4. getblockhash

Returns the hash value of the corresponding block according to the specified index.

#### Parameter instruction

Index: block index.

#### Example

Request:

```
{
  "jsonrpc": "2.0",
  "method": "getblockhash",
  "params": [10000],
  "id": 1
}
```

Reponse:

```
{
  "desc":"SUCCESS",
  "error":0,
  "jsonrpc": "2.0",
  "id": 1,
  "result": "4c1e879872344349067c3b1a30781eeb4f9040d3795db7922f513f6f9660b9b2"
}
```

#### 5. getconnectioncount

Get the current number of connections for the node.

#### Example

Request:

```
{
  "jsonrpc": "2.0",
  "method": "getconnectioncount",
  "params": [],
  "id": 1
}
```

Response:

```
{
  "desc":"SUCCESS",
  "error":0,
  "jsonrpc": "2.0",
  "id": 1,
  "result": 10
}
```

#### 6. getgenerateblocktime

Get a list of unconfirmed transactions in memory.

#### Example

Request:

```
{
  "jsonrpc": "2.0",
  "method": "getgenerateblocktime",
  "params": [],
  "id": 1
}
```

Reponse:

```
{
  "desc":"SUCCESS",
  "error":0,
  "jsonrpc": "2.0",
  "id": 1,
  "result": [
    "b4534f6d4c17cda008a76a1968b7fa6256cd90ca448739eae8e828698ccc44e7"
  ]
}
```

These are the undetermined transactions received by the node, that is, those with zero confirmed transactions.

#### 7. getrawtransaction

Returns the corresponding transaction information based on the specified hash value.

#### Parameter instruction

txid: transaction ID

Verbose: Optional parameter, the default value of verbose is 0, when verbose is 0, it returns the block serialized information, which is represented by a hexadecimal string. To get detailed information from it, you need to call the SDK to deserialize. When verbose is 1, the detailed information of the corresponding block is returned, which is represented by a JSON format string.

#### Example

Request:

```
{
  "jsonrpc": "2.0",
  "method": "getrawtransaction",
  "params": ["f4250dab094c38d8265acc15c366dc508d2e14bf5699e12d9df26577ed74d657"],
  "id": 1
}
```

Response:

```
{
  "desc":"SUCCESS",
  "error":0,
  "jsonrpc": "2.0",
  "id": 1,
  "result": "80000001195876cb34364dc38b730077156c6bc3a7fc570044a66fbfeeea56f71327e8ab0000029b7cffdaa674beae0f930ebe6085af9093e5fe56b34a5c220ccdcf6efc336fc500c65eaf440000000f9a23e06f74cf86b8827a9108ec2e0f89ad956c9b7cffdaa674beae0f930ebe6085af9093e5fe56b34a5c220ccdcf6efc336fc50092e14b5e00000030aab52ad93f6ce17ca07fa88fc191828c58cb71014140915467ecd359684b2dc358024ca750609591aa731a0b309c7fb3cab5cd0836ad3992aa0a24da431f43b68883ea5651d548feb6bd3c8e16376e6e426f91f84c58232103322f35c7819267e721335948d385fae5be66e7ba8c748ac15467dcca0693692dac"
}
or
{
  "desc":"SUCCESS",
  "error":0,
  "jsonrpc": "2.0",
  "id": 1,
  "result": {
    "txid": "f4250dab094c38d8265acc15c366dc508d2e14bf5699e12d9df26577ed74d657",
    "size": 262,
    "type": "ContractTransaction",
    "version": 0,
    "attributes": [],
    "vin": [
      {
        "txid": "abe82713f756eaeebf6fa6440057fca7c36b6c157700738bc34d3634cb765819",
        "vout": 0
      }
    ],
    "vout": [
      {
        "n": 0,
        "asset": "c56f33fc6ecfcd0c225c4ab356fee59390af8560be0e930faebe74a6daff7c9b",
        "value": "2950",
        "address": "AHCNSDkh2Xs66SzmyKGdoDKY752uyeXDrt"
      },
      {
        "n": 1,
        "asset": "c56f33fc6ecfcd0c225c4ab356fee59390af8560be0e930faebe74a6daff7c9b",
        "value": "4050",
        "address": "ALDCagdWUVV4wYoEzCcJ4dtHqtWhsNEEaR"
      }
    ],
    "sys_fee": "0",
    "net_fee": "0",
    "scripts": [
      {
        "invocation": "40915467ecd359684b2dc358024ca750609591aa731a0b309c7fb3cab5cd0836ad3992aa0a24da431f43b68883ea5651d548feb6bd3c8e16376e6e426f91f84c58",
        "verification": "2103322f35c7819267e721335948d385fae5be66e7ba8c748ac15467dcca0693692dac"
      }
    ],
    "blockhash": "9c814276156d33f5dbd4e1bd4e279bb4da4ca73ea7b7f9f0833231854648a72c",
    "confirmations": 144,
    "blocktime": 1496719422
  }
}
```

#### 8. sendrawtransaction

Broadcast transaction.

#### Parameter instruction

Hex: Serialized signed transactions constructed in the program into hexadecimal strings.

#### Example

Request:

```
{
  "jsonrpc": "2.0",
  "method": "sendrawtransaction",
  "params": ["80000001195876cb34364dc38b730077156c6bc3a7fc570044a66fbfeeea56f71327e8ab0000029b7cffdaa674beae0f930ebe6085af9093e5fe56b34a5c220ccdcf6efc336fc500c65eaf440000000f9a23e06f74cf86b8827a9108ec2e0f89ad956c9b7cffdaa674beae0f930ebe6085af9093e5fe56b34a5c220ccdcf6efc336fc50092e14b5e00000030aab52ad93f6ce17ca07fa88fc191828c58cb71014140915467ecd359684b2dc358024ca750609591aa731a0b309c7fb3cab5cd0836ad3992aa0a24da431f43b68883ea5651d548feb6bd3c8e16376e6e426f91f84c58232103322f35c7819267e721335948d385fae5be66e7ba8c748ac15467dcca0693692dac"],
  "id": 1
}
```

Response:

```
{
  "desc":"SUCCESS",
  "error":0,
  "jsonrpc": "2.0",
  "id": 1,
  "result": false
}
```

Response instruction:

When result is true, the current transaction broadcast is successful.

When result is false, it means that the current transaction broadcast failed because of double costs, incomplete signatures, etc.

In this example, a confirmed transaction was broadcast, the broadcast failed because of double costs.

#### 9. getstorage

Returns the stored value according to the contract script hashes and stored key.

#### Parameter instruction

script\_hash: Contract script hash.

Key: stored key \(required to be converted into hex string\)

#### Example

Request:

```
{
    "jsonrpc": "2.0",
    "method": "getstorage",
    "params": ["03febccf81ac85e3d795bc5cbd4e84e907812aa3", "5065746572"],
    "id": 15
}
```

Response:

```
{
    "desc":"SUCCESS",
    "error":0,
    "jsonrpc": "2.0",
    "id": 15,
    "result": "4c696e"
}
```

#### 10. getversion

Get the version information of the query node.

#### Example

Request:

```
{
  "jsonrpc": "2.0",
  "method": "getversion",
  "params": [],
  "id": 3
}
```

Response:

```
{
  "desc":"SUCCESS",
  "error":0,
  "jsonrpc": "2.0",
  "id": 3,
  "result": {
      "port": 0,
      "nonce": 156443862,
      "useragent": "/ONT:1.0.0/"
  }
}
```

#### 11. getsmartcodeevent

Get smartcode event.

#### Parameter instruction

blockheight: getsmartcodeevent by blockheight
or
txHash: getsmartcodeevent by blockheight

#### Example

Request:

```
{
  "jsonrpc": "2.0",
  "method": "getsmartcodeevent",
  "params": [3],
  "id": 3
}
```
or
```
{
  "jsonrpc": "2.0",
  "method": "getsmartcodeevent",
  "params": ["3ba4b4e463a717635614595378f2aac78feacc7d4dfda075bfcf9328cbbcdb7c"],
  "id": 3
}
```
Response:

```
{
  "desc":"SUCCESS",
  "error":0,
  "jsonrpc": "2.0",
  "id": 3,
  "result": {

  }
}
```

#### 12. getblocksysfee

According to the specified index, return the system fee before the block.

#### Parameter instruction

Index: Block index

#### Example

Request:

```
{
  "jsonrpc": "2.0",
  "method": "getblocksysfee",
  "params": [1005434],
  "id": 1
}
```

Response:

```
{
    "desc":"SUCCESS",
    "error":0,
    "jsonrpc": "2.0",
    "id": 1,
    "result": "195500"
}
```

Response instruction:

Result: The system fee before the block and the unit is OntGas.

#### 13. getcontractstate

According to the contract script hash, query the contract information.

#### Parameter instruction

script\_hash: Contract script hash.

#### Example

Request:

```
{
  "jsonrpc": "2.0",
  "method": "getcontractstate",
  "params": ["8a4d2865d01ec8e6add72e3dfdd20c12f44834e3"],
  "id": 1
}
```

Response:

```
{
    "desc":"SUCCESS",
    "error":0,
    "jsonrpc": "2.0",
    "id": 1,
    "result": {
        "version": 0,
        "code": {
            "hash": "8a4d2865d01ec8e6add72e3dfdd20c12f44834e3",
            "script": "746b4c04000000004c04000000004c04000000004c04000000004c04000000004c04000000004c04000000004c04000000004c04000000004c04000000004c04000000004c04000000004c04000000004c04000000004c04000000004c04000000004c04000000004c04000000004c04000000004c04000000004c04000000004c04000000004c04000000004c04000000004c040000000061744c0403000000936c766b9479744c0406000000936c766b9479617cac744c0406000000948c6c766b947275744c0406000000948c6c766b9479641b004c0401000000744c0407000000948c6c766b94727562b207744c0400000000936c766b9479744c0406000000936c766b9479617cac4c04000000009c744c0408000000948c6c766b947275744c0408000000948c6c766b9479641b004c0400000000744c0407000000948c6c766b947275625607744c0404000000936c766b9479744c0409000000948c6c766b947275744c0409000000948c6c766b947964400061744c0401000000936c766b9479744c0400000000948c6c766b947275744c0402000000936c766b9479744c0401000000948c6c766b94727561623d0061744c0402000000936c766b9479744c0400000000948c6c766b947275744c0401000000936c766b9479744c0401000000948c6c766b947275614c0400000000744c0402000000948c6c766b9472754c0400000000744c0403000000948c6c766b94727561682953797374656d2e457865637574696f6e456e67696e652e476574536372697074436f6e7461696e6572616823416e745368617265732e5472616e73616374696f6e2e4765745265666572656e636573744c0404000000948c6c766b94727561744c0404000000948c6c766b9479744c040a000000948c6c766b9472754c0400000000744c040b000000948c6c766b947275629501744c040a000000948c6c766b9479744c040b000000948c6c766b9479c3744c040c000000948c6c766b94727561744c040c000000948c6c766b947961681e416e745368617265732e4f75747075742e4765745363726970744861736861682953797374656d2e457865637574696f6e456e67696e652e476574456e7472795363726970744861736887744c040d000000948c6c766b947275744c040d000000948c6c766b947964c70061744c040c000000948c6c766b947961681b416e745368617265732e4f75747075742e47657441737365744964744c0400000000948c6c766b9479874c04000000009c744c040e000000948c6c766b947275744c040e000000948c6c766b9479641b004c0400000000744c0407000000948c6c766b94727562cd04744c0402000000948c6c766b9479744c040c000000948c6c766b9479616819416e745368617265732e4f75747075742e47657456616c756593744c0402000000948c6c766b9472756161744c040b000000948c6c766b94794c040100000093744c040b000000948c6c766b947275744c040b000000948c6c766b9479744c040a000000948c6c766b9479c09f6350fe61682953797374656d2e457865637574696f6e456e67696e652e476574536372697074436f6e7461696e6572616820416e745368617265732e5472616e73616374696f6e2e4765744f757470757473744c0405000000948c6c766b94727561744c0405000000948c6c766b9479744c040f000000948c6c766b9472754c0400000000744c0410000000948c6c766b947275621c02744c040f000000948c6c766b9479744c0410000000948c6c766b9479c3744c0411000000948c6c766b94727561744c0411000000948c6c766b947961681e416e745368617265732e4f75747075742e4765745363726970744861736861682953797374656d2e457865637574696f6e456e67696e652e476574456e7472795363726970744861736887744c0412000000948c6c766b947275744c0412000000948c6c766b9479644e0161744c0411000000948c6c766b947961681b416e745368617265732e4f75747075742e47657441737365744964744c0400000000948c6c766b947987744c0413000000948c6c766b947275744c0413000000948c6c766b9479644e00744c0402000000948c6c766b9479744c0411000000948c6c766b9479616819416e745368617265732e4f75747075742e47657456616c756594744c0402000000948c6c766b94727562a600744c0411000000948c6c766b947961681b416e745368617265732e4f75747075742e47657441737365744964744c0401000000948c6c766b947987744c0414000000948c6c766b947275744c0414000000948c6c766b9479644b00744c0403000000948c6c766b9479744c0411000000948c6c766b9479616819416e745368617265732e4f75747075742e47657456616c756593744c0403000000948c6c766b9472756161744c0410000000948c6c766b94794c040100000093744c0410000000948c6c766b947275744c0410000000948c6c766b9479744c040f000000948c6c766b9479c09f63c9fd744c0402000000948c6c766b94794c0400000000a1744c0415000000948c6c766b947275744c0415000000948c6c766b9479641b004c0401000000744c0407000000948c6c766b947275622301744c0404000000936c766b9479744c0416000000948c6c766b947275744c0416000000948c6c766b947964720061744c0403000000948c6c766b94794c0400e1f50595744c0402000000948c6c766b9479744c0405000000936c766b9479959f744c0417000000948c6c766b947275744c0417000000948c6c766b9479641b004c0400000000744c0407000000948c6c766b947275628b0061626f0061744c0402000000948c6c766b94794c0400e1f50595744c0403000000948c6c766b9479744c0405000000936c766b947995a0744c0418000000948c6c766b947275744c0418000000948c6c766b9479641b004c0400000000744c0407000000948c6c766b947275621c00614c0401000000744c0407000000948c6c766b947275620300744c0407000000948c6c766b947961748c6c766b946d748c6c766b946d748c6c766b946d748c6c766b946d748c6c766b946d748c6c766b946d748c6c766b946d748c6c766b946d748c6c766b946d748c6c766b946d748c6c766b946d748c6c766b946d748c6c766b946d748c6c766b946d748c6c766b946d748c6c766b946d748c6c766b946d748c6c766b946d748c6c766b946d748c6c766b946d748c6c766b946d748c6c766b946d748c6c766b946d748c6c766b946d748c6c766b946d746c768c6b946d746c768c6b946d746c768c6b946d746c768c6b946d746c768c6b946d746c768c6b946d746c768c6b946d6c7566",
            "parameters": [
                "Hash160",
                "Hash256",
                "Hash256",
                "Hash160",
                "Boolean",
                "Integer",
                "Signature"
            ],
            "returntype": "Boolean"
        },
        "storage": false,
        "name": "AgencyContract",
        "code_version": "2.0.1-preview1",
        "author": "Erik Zhang",
        "email": "erik@antshares.org",
        "description": "Agency Contract 2.0"
    }
}
```

#### 14. getmempooltxstate

Query the transaction status in the memory pool.

#### Parameter instruction

tx\_hash: transaction hash.

#### Example

Request:

```
{
  "jsonrpc": "2.0",
  "method": "getmempooltxstate",
  "params": ["773dd2dae4a9c9275290f89b56e67d7363ea4826dfd4fc13cc01cf73a44b0d0e"],
  "id": 1
}
```

Response:

```
{
    "desc":"SUCCESS",
    "error":0,
    "jsonrpc": "2.0",
    "id": 1,
    "result": {
    }
}
```

#### 15. getsmartcodeevent

Get smartcontract event.

#### Parameter instruction

Height: block height.

#### Example

Request:

```
{
  "jsonrpc": "2.0",
  "method": "getsmartcodeevent",
  "params": [101],
  "id": 1
}
```

Response:

```
{
    "desc":"SUCCESS",
    "error":0,
    "jsonrpc": "2.0",
    "id": 1,
    "result": {
    }
}
```
#### 15. getblockheightbytxhash
get blockheight by txhash
#### Parameter instruction
txhash: transaction hash
#### Example

Request:

```
{
  "jsonrpc": "2.0",
  "method": "getblocksysfee",
  "params": ["TA5uYzLU2vBvvfCMxyV2sdzc9kPqJzGZWq"],
  "id": 1
}
```

#### 16. getbalance

return balance of base58 account address.

#### Parameter instruction

address: Base58-encoded form of account address

#### Example

Request:

```
{
  "jsonrpc": "2.0",
  "method": "getblocksysfee",
  "params": ["TA5uYzLU2vBvvfCMxyV2sdzc9kPqJzGZWq"],
  "id": 1
}
```

Response:

```
{
   "desc":"SUCCESS",
   "error":0,
   "id":1,
   "jsonpc":"2.0",
   "result":{
       "ont":"24999862561046528",
       "ong":"0"
       }
}
```

Response instruction:

Result: The system fee before the block and the unit is OntGas.

## Error code

errorcode instruction

| Field | Type | Description |
| :--- | :--- | :--- |
| 0 | int64 | SUCCESS |
| 41001 | int64 | SESSION\_EXPIRED: invalided or expired session |
| 41002 | int64 | SERVICE\_CEILING: reach service limit |
| 41003 | int64 | ILLEGAL\_DATAFORMAT: illegal dataformat |
| 41004 | int64 | INVALID\_VERSION: invalid version |
| 42001 | int64 | INVALID\_METHOD: invalid method |
| 42002 | int64 | INVALID\_PARAMS: invalid params |
| 43001 | int64 | INVALID\_TRANSACTION: invalid transaction |
| 43002 | int64 | INVALID\_ASSET: invalid asset |
| 43003 | int64 | INVALID\_BLOCK: invalid block |
| 44001 | int64 | UNKNOWN\_TRANSACTION: unknown transaction |
| 44002 | int64 | UNKNOWN\_ASSET: unknown asset |
| 44003 | int64 | UNKNOWN\_BLOCK: unknown block |
| 45001 | int64 | INTERNAL\_ERROR: internel error |
| 47001 | int64 | SMARTCODE\_ERROR: smartcode error |



