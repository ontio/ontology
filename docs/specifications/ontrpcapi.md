# ONT Rpc Api

* [Introduction](#Introduction)
* [Rpc api list](#Rpc api list)
* [Error code](#Error code)

## Introduction

Request parameter description:

| Field | Type | Description |
| :---| :---| :---|
| jsonrpc | string | jsonrpc version |
| method | string | method name |
| params | string | method required parameters |
| id | int | any value |

Response parameter description:

| Field | Type | Description |
| :---| :---| :---|
| desc| string | resopnse description |
| error | int64 | error code |
| jsonrpc | string | jsonrpc version |
| id | int | any value |
| result | object/string/bool | program execution result |

Note: The type of result varies with the request.

Block field description

| Field | Type | Description |
| :--- | :--- | :--- |
| Header | *Header |  |
| Transactions | []*Transaction ||
| hash | *Uint256 | |

Header field description

| Field | Type | Description |
| :--- | :--- | :--- |
| Version | uint32 | version number |
| PrevBlockHash | Uint256 | The hash of the previous block |
| TransactionsRoot | Uint256 | The root of the Merkle tree for all transactions in this block |
| BlockRoot | Uint256 | blockroot |
| Timestamp | int | block timestamp,uinix timestamp |
| Height | int | block height |
| ConsensusData | uint64 |  |
| NextBookKeeper | Address | Accounting contract hash value for the next block |
| BookKeepers | []*crypto.PubKey ||
| SigData | [][]byte ||
| Hash | Uint256 | Script to verify the block |

Transaction field description

| Field | Type | Description |
| :--- | :--- | :--- |
| Version| byte | version number |
| TxType | TransactionType | transaction type |
| Payload | Payload | payload |
| Nonce | uint32 | random number |
| Attributes | []*TxAttribute |  |
| Fee | []*Fee | transaction fees  |
| NetworkFee | Fixed64 | neitwork fees |
| Sigs | []*Sig | signature array |
| Hash | *Uint256 | transaction hash |

## Rpc api list

| Method | Parameters | Description | Note |
| :---| :---| :---| :---|
| getbestblockhash |  | get the hash of the highest height block in the main chain |  |
| getblock | height or blockhash,[verbose] | get block by block height or block hash | verbose can be 0 or 1,response is different |
| getblockcount |  | get the number of blocks |  |
| getblockhash | height | get block hash by block height |  |
| getconnectioncount|  | get the current number of connections for the node |  |
| getgenerateblocktime|  | The time required to create a new block |  |
| getrawtransaction | transactionhash | Returns the corresponding transaction information based on the specified hash value. |  |
| sendrawtransaction | hex | Broadcast transaction. | Serialized signed transactions constructed in the program into hexadecimal strings |
| getstorage | script_hash | Returns the stored value according to the contract script hashes and stored key. |  |
| getversion |  | Get the version information of the query node |  |
| getblocksysfee |  | According to the specified index, return the system fee before the block. |  |
| getcontractstate | script_hash | According to the contract script hash, query the contract information. |  |
| getmempooltxstate | tx_hash | Query the transaction status in the memory pool. |  |
| getsmartcodeevent |  | Get smartcode event |  |
| getblockheightbytxhash | tx_hash | get blockheight of txhash|  |
| getbalance | address | return balance of base58 account address. |  |


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
        "Hash": "95555da65d6feaa7cde13d6bf12131f750b670569d98c63813441cf24a99c0d2",
        "Header": {
            "Version": 0,
            "PrevBlockHash": "205c905493c7c1e3be7cd58542e45aafb007edcb8363f8ff555f63745f1b7ce5",
            "TransactionsRoot": "4452db2634d81e80048002c2f327b25ded4e547ebfcc1b28f28608938b9d2154",
            "BlockRoot": "42e01a2b27c182d4e115883c3b166a0fbc019efe2498b568b7febcc83a35346e",
            "Timestamp": 1522295648,
            "Height": 2,
            "ConsensusData": 10322907760044199803,
            "NextBookkeeper": "TAAr9AH4NqxXSKur7XTUbmP8wsKD4KPL2t",
            "Bookkeepers": [
                "120203e45fe0189a36b284e6080c6983cf12879d239886ecee1e257ab992970ecaa000"
            ],
            "SigData": [
                "014ed021011a6e0a4e9771b0be9fd156f9fc411968ce1dc4aed18382c85f6827d50373f3e3931966066cdc7dfab52823b79c80df8af25569c33ddf8140df5385b6"
            ],
            "Hash": "95555da65d6feaa7cde13d6bf12131f750b670569d98c63813441cf24a99c0d2"
        },
        "Transactions": [
            {
                "Version": 0,
                "Nonce": 0,
                "TxType": 0,
                "Payload": {
                    "Nonce": 1522295648487066000
                },
                "Attributes": [],
                "Fee": [],
                "NetworkFee": 0,
                "Sigs": [
                    {
                        "PubKeys": [
                            "120203e45fe0189a36b284e6080c6983cf12879d239886ecee1e257ab992970ecaa000"
                        ],
                        "M": 1,
                        "SigData": [
                            "01021197ad4140a50442b700ad814aeb2595578bf4d97e187a69aacf35917be4a27f76bc1dad2ee9bb386be79ca9638e78e14c869edbc3556499b06cc9c9b9452e"
                        ]
                    }
                ],
                "Hash": "4452db2634d81e80048002c2f327b25ded4e547ebfcc1b28f28608938b9d2154"
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

```

or

```
{
    "desc": "SUCCESS",
    "error": 0,
    "id": 1,
    "jsonpc": "2.0",
    "result": {
        "Version": 0,
        "Nonce": 3377520203,
        "TxType": 209,
        "Payload": {
            "Code": "00ff00000000000000000000000000000000000001087472616e736665722d000100017d439492af400d014c2b0cc4975d7252868d8001c484de9cde9d10c3bf49362e6d66a6c3b196b70164",
            "GasLimit": 0,
            "VmType": 255
        },
        "Attributes": [
            {
                "Usage": 0,
                "Data": "34336234663163352d373764392d346634342d626262662d326539396136656538376237"
            }
        ],
        "Fee": [
            {
                "Amount": 0,
                "Payer": "017d439492af400d014c2b0cc4975d7252868d80"
            }
        ],
        "NetworkFee": 0,
        "Sigs": [
            {
                "PubKeys": [
                    "12020206b47806887dfb13679ae884e7843ef263f54a861792502100f6bb3f5bd896cc"
                ],
                "M": 1,
                "SigData": [
                    "012a0623b31b681c74866c9e72c255ac026a1fcc61867b3f1dc7a25266939e73a24c87c2aceda41174b85a872b11dbf7020a4d52dffbbfefdb704406738dd042bf"
                ]
            }
        ],
        "Hash": "a724c0215afa1aeb31be857f2fc69038cf557b4748941bfed8281473b39152e7"
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

or

```
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
> result: Hexadecimal string

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
txHash: getsmartcodeevent by txhash

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
  "params": ["fff49c809d302a2956e9dc0012619a452d4b846c"],
  "id": 1
}
```

Response:

```
{
    "desc": "SUCCESS",
    "error": 0,
    "id": 1,
    "jsonpc": "2.0",
    "result": {
        "VmType": 255,
        "Code": "4f4e5420546f6b656e",
        "NeedStorage": true,
        "Name": "ONT",
        "CodeVersion": "1.0",
        "Author": "Ontology Team",
        "Email": "contact@ont.io",
        "Description": "Ontology Network ONT Token"
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

#### 15. getblockheightbytxhash
get blockheight by txhash
#### Parameter instruction
txhash: transaction hash
#### Example

Request:

```
{
  "jsonrpc": "2.0",
  "method": "getblockheightbytxhash",
  "params": ["c453557af780fe403db6e954ebc9adeafd5818c596c6c60e5cc42851c5b41884"],
  "id": 1
}
```

Response:
```
{
    "desc": "SUCCESS",
    "error": 0,
    "id": 1,
    "jsonpc": "2.0",
    "result": 10
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
  "method": "getbalance",
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
