# Ontology Websocket API

[English](websocket_api.md)|中文

* [介绍](#介绍)
* [Websocket接口列表](#websocket接口列表)
* [错误代码](#错误代码)

## 介绍

本文档是Ontology的websocket接口文档，详细定义了各个接口所需的参数与返回值。

### 响应参数定义

| Field | Type | Description |
| :--- | :--- | :--- |
| Action | string | 响应动作名称 |
| Desc | string | 响应结果描述 |
| Error | int64 | 错误代码 |
| Result | object | 执行结果 |
| Version | string | 版本号 |
| Id | int64 | 请求id|

## Websocket接口列表

| Method | Parameter | Description |
| :---| :---| :---|
| [heartbeat](#1-heartbeat) |  | 发送心跳信号 |
| [subscribe](#2-subscribe) | [ContractsFilter],[SubscribeEvent],[SubscribeJsonBlock],[SubscribeRawBlock],[SubscribeBlockTxHashs] | 订阅某个服务 |
| [getconnectioncount](#3-getconnectioncount) |  | 得到当前连接的节点数量 |
| [getblocktxsbyheight](#4-getblocktxsbyheight) | height | 返回对应高度的区块中落账的所有交易哈希 |
| [getblockbyheight](#5-getblockbyheight) | height | 得到该高度的区块的详细信息 |
| [getblockbyhash](#6-getblockbyhash) | hash | 通过区块哈希得到区块信息 |
| [getblockheight](#7-getblockheight) |  | 得到当前网络上的区块高度 |
| [getblockhash](#8-getblockhash) | height | 根据高度得到对应区块的哈希 |
| [gettransaction](#9-gettransaction) | hash,[raw] | 通过交易哈希得到该交易的信息 |
| [sendrawtransaction](#10-sendrawtransaction) | data,[PreExec] | 向ontology网络发送交易, 如果 preExec=1，则交易为预执行 |
| [getstorage](#11-getstorage) | hash,key | 通过合约地址哈希和键得到对应的值 |
| [getbalance](#12-getbalance) | address | 得到该地址的账户的余额 |
| [getcontract](#13-getcontract) | hash | 根据合约地址哈希得到合约信息 |
| [getsmartcodeeventbyheight](#14-getsmartcodeeventbyheight) | height | 得到该高度区块上的智能合约执行结果 |
| [getsmartcodeeventbyhash](#15-getsmartcodeeventbyhash) | hash | 通过交易哈希得到该交易的执行结果 |
| [getblockheightbytxhash](#16-getblockheightbytxhash) | hash | 通过交易哈希得到该交易落账的区块高度 |
| [getmerkleproof](#17-getmerkleproof) | hash | 通过交易哈希得到该交易的merkle证明 |
| [getsessioncount](#18-getsessioncount) |  | 得到会话数量 |
| [getgasprice](#19-getgasprice) |  | 得到gas的价格 |
| [getallowance](#20-getallowance) | asset, from, to | 返回允许从from账户转出到to账户的额度 |
| [getunboundong](#21-getunboundong) | address | 返回该账户未提取的ong数量 |
| [getmempooltxstate](#22-getmempooltxstate) | hash | 通过交易哈希得到内存中该交易的状态 |
| [getmempooltxcount](#23-getmempooltxcount) |  | 得到内存中的交易的数量 |
| [getversion](#24-getversion) |  | 得到版本信息 |
| [getnetworkid](#25-getnetworkid) |  | 得到network id |
| [getgrantong](#26-getgrantong) |  | 得到grant ong |

###  1. heartbeat

如果超过五分钟没有发送心跳信号，则连接关闭。

#### Request Example:

```
{
    "Action": "heartbeat",
    "Id":12345, //optional
    "Version": "1.0.0"
}
```

#### Response example:

```
{
    "Action": "heartbeat",
    "Desc": "SUCCESS",
    "Error": 0,
    "Result": {
        "SubscribeEvent":false,
        "SubscribeJsonBlock":false,
        "SubscribeRawBlock":false,
        "SubscribeBlockTxHashs":false
    }
    "Version": "1.0.0"
}
```

###  2. subscribe
订阅某个服务。

#### Request Example:

```
{
    "Action": "subscribe",
    "Version": "1.0.0",
    "Id":12345, //optional
    "ContractsFilter":["ecceb5863d20b9d05412a5f2641167e716628932"], //optional
    "SubscribeEvent":false, //optional
    "SubscribeJsonBlock":true, //optional
    "SubscribeRawBlock":false, //optional
    "SubscribeBlockTxHashs":false //optional
}
```

#### Response example:

```
{
    "Action": "subscribe",
    "Desc": "SUCCESS",
    "Error": 0,
    "Result": {
        "ContractsFilter":["ecceb5863d20b9d05412a5f2641167e716628932"],
        "SubscribeEvent":false,
        "SubscribeJsonBlock":true,
        "SubscribeRawBlock":false,
        "SubscribeBlockTxHashs":false
    }
    "Version": "1.0.0"
}
```


### 3. getconnectioncount

得到当前连接的节点数量。


#### Request Example:

```
{
    "Action": "getconnectioncount",
    "Id":12345, //optional
    "Version": "1.0.0"
}
```

#### Response Example:

```
{
    "Action": "getconnectioncount",
    "Desc": "SUCCESS",
    "Error": 0,
    "Result": 4,
    "Version": "1.0.0"
}
```
### 4. getblocktxsbyheight

返回对应高度的区块中落账的所有交易哈希。


#### Request Example:

```
{
    "Action": "getblocktxsbyheight",
    "Version": "1.0.0",
    "Id":12345, //optional
    "Height": 100
}
```

#### Response Example:

```
{
    "Action": "getblocktxsbyheight",
    "Desc": "SUCCESS",
    "Error": 0,
    "Result": {
        "Hash": "ea5e5219d2f1591f4feef89885c3f38c83d3a3474a5622cf8cd3de1b93849603",
        "Height": 100,
        "Transactions": [
            "37e017cb9de93aa93ef817e82c555812a0a6d5c3f7d6c521c7808a5a77fc93c7"
        ]
    },
    "Version": "1.0.0"
}
```
### 5. getblockbyheight

得到该高度的区块的详细信息。

raw：可选参数，默认值为零，不设置时为默认值。当值为1时，接口返回区块序列化后的信息，该信息以十六进制字符串表示。如果要得到区块的具体信息，需要调用
 SDK中的方法对该字符串进行反序列化。当值为0时，将以json格式返回对应区块的详细信息。

#### Request Example:

```
{
    "Action": "getblockbyheight",
    "Version": "1.0.0",
    "Id":12345, //optional
    "Raw": "0",
    "Height": 100
}
```

#### Response Example:

```
{
    "Action": "getblockbyheight",
    "Desc": "SUCCESS",
    "Error": 0,
    "Result": {
        "Hash": "ea5e5219d2f1591f4feef89885c3f38c83d3a3474a5622cf8cd3de1b93849603",
        "Header": {
            "Version": 0,
            "PrevBlockHash": "fc3066adb581c5aee8edaa47eecda2b7cc039c8662757f8b1e3c3aed60314353",
            "TransactionsRoot": "37e017cb9de93aa93ef817e82c555812a0a6d5c3f7d6c521c7808a5a77fc93c7",
            "BlockRoot": "7154a6dcb3c23254334bc1f5d8f054c143a39ff28f46fdeb8a9c7488147ccec6",
            "Timestamp": 1522313652,
            "Height": 100,
            "ConsensusData": 18012644264110396442,
            "NextBookkeeper": "TABrSU6ABhj6Rdw5KozV53wvZNSUATgKHW",
            "Bookkeepers": [
                "120203fe4f9ba2022b68595dd163f4a92ac80f918919674de2d6e2a7e04a10c59d0066"
            ],
            "SigData": [
                "01a2369280b0ff75bed85f351d3ef0dd58add118328c1ed2f7d3320df32cb4bd55541f1bb8e11ad093bd24da3de4cd12464800310bfdb49dc62d42d97ca0549762"
            ],
            "Hash": "ea5e5219d2f1591f4feef89885c3f38c83d3a3474a5622cf8cd3de1b93849603"
        },
        "Transactions": [
            {
                "Version": 0,
                "Nonce": 0,
                "TxType": 0,
                "Payload": {
                    "Nonce": 1522313652068190000
                },
                "Attributes": [],
                "Fee": [],
                "NetworkFee": 0,
                "Sigs": [
                    {
                        "PubKeys": [
                            "120203fe4f9ba2022b68595dd163f4a92ac80f918919674de2d6e2a7e04a10c59d0066"
                        ],
                        "M": 1,
                        "SigData": [
                            "017d3641607c894dd85f455c71a94afaea2661acbe372ff8f3f4c7921b0c768756e3a6e9308a4c4c8b1b58e717f1486a2f10f5bc809b803a27c10a2cd579778a54"
                        ]
                    }
                ],
                "Hash": "37e017cb9de93aa93ef817e82c555812a0a6d5c3f7d6c521c7808a5a77fc93c7"
            }
        ]
    },
    "Version": "1.0.0"
}
```
### 6. getblockbyhash

通过区块哈希得到区块信息。

raw：可选参数，默认值为零，不设置时为默认值。当值为1时，接口返回区块序列化后的信息，该信息以十六进制字符串表示。如果要得到区块的具体信息，需要调用
 SDK中的方法对该字符串进行反序列化。当值为0时，将以json格式返回对应区块的详细信息。

#### Request Example:

```
{
    "Action": "getblockbyhash",
    "Version": "1.0.0",
    "Id":12345, //optional
    "Raw": "0",
    "Hash": "7c3e38afb62db28c7360af7ef3c1baa66aeec27d7d2f60cd22c13ca85b2fd4f3"
}
```

#### Response Example:

```
{
    "Action": "getblockbyhash",
    "Desc": "SUCCESS",
    "Error": 0,
    "Result": {
        "Hash": "ea5e5219d2f1591f4feef89885c3f38c83d3a3474a5622cf8cd3de1b93849603",
        "Header": {
            "Version": 0,
            "PrevBlockHash": "fc3066adb581c5aee8edaa47eecda2b7cc039c8662757f8b1e3c3aed60314353",
            "TransactionsRoot": "37e017cb9de93aa93ef817e82c555812a0a6d5c3f7d6c521c7808a5a77fc93c7",
            "BlockRoot": "7154a6dcb3c23254334bc1f5d8f054c143a39ff28f46fdeb8a9c7488147ccec6",
            "Timestamp": 1522313652,
            "Height": 100,
            "ConsensusData": 18012644264110396442,
            "NextBookkeeper": "TABrSU6ABhj6Rdw5KozV53wvZNSUATgKHW",
            "Bookkeepers": [
                "120203fe4f9ba2022b68595dd163f4a92ac80f918919674de2d6e2a7e04a10c59d0066"
            ],
            "SigData": [
                "01a2369280b0ff75bed85f351d3ef0dd58add118328c1ed2f7d3320df32cb4bd55541f1bb8e11ad093bd24da3de4cd12464800310bfdb49dc62d42d97ca0549762"
            ],
            "Hash": "ea5e5219d2f1591f4feef89885c3f38c83d3a3474a5622cf8cd3de1b93849603"
        },
        "Transactions": [
            {
                "Version": 0,
                "Nonce": 0,
                "TxType": 0,
                "Payload": {
                    "Nonce": 1522313652068190000
                },
                "Attributes": [],
                "Fee": [],
                "NetworkFee": 0,
                "Sigs": [
                    {
                        "PubKeys": [
                            "120203fe4f9ba2022b68595dd163f4a92ac80f918919674de2d6e2a7e04a10c59d0066"
                        ],
                        "M": 1,
                        "SigData": [
                            "017d3641607c894dd85f455c71a94afaea2661acbe372ff8f3f4c7921b0c768756e3a6e9308a4c4c8b1b58e717f1486a2f10f5bc809b803a27c10a2cd579778a54"
                        ]
                    }
                ],
                "Hash": "37e017cb9de93aa93ef817e82c555812a0a6d5c3f7d6c521c7808a5a77fc93c7"
            }
        ]
    },
    "Version": "1.0.0"
}
```

### 7. getblockheight

得到当前网络上的区块高度。


#### Request Example:

```
{
    "Action": "getblockheight",
    "Id":12345, //optional
    "Version": "1.0.0"
}
```


#### Response Example:

```
{
    "Action": "getblockheight",
    "Desc": "SUCCESS",
    "Error": 0,
    "Result": 327,
    "Version": "1.0.0"
}
```

### 8. getblockhash

根据高度得到对应区块的哈希。


#### Request Example:

```
{
    "Action": "getblockhash",
    "Id":12345, //optional
    "Version": "1.0.0",
    "Height": 100
}
```

#### Response Example:

```
{
    "Action": "getblockhash",
    "Desc": "SUCCESS",
    "Error": 0,
    "Result": "3b90ddc4d33c4954c3d87736120e94915f963546861987757f358c9376422255",
    "Version": "1.0.0"
}
```

### 9. gettransaction

通过交易哈希得到该交易的信息。

raw：可选参数，默认值为零，不设置时为默认值。当值为1时，接口返回交易序列化后的信息，该信息以十六进制字符串表示。如果要得到交易的具体信息，需要调用
 SDK中的方法对该字符串进行反序列化。当值为0时，将以json格式返回对应交易的详细信息。

#### Request Example:

```
{
    "Action": "gettransaction",
    "Version": "1.0.0",
    "Id":12345, //optional
    "Hash": "3b90ddc4d33c4954c3d87736120e94915f963546861987757f358c9376422255",
    "Raw": "0"
}
```
#### Response Example:

```
{
    "Action": "gettransaction",
    "Desc": "SUCCESS",
    "Error": 0,
    "Result": {
        "Version": 0,
        "Nonce": 3743545316,
        "GasPrice": 500,
        "GasLimit": 20000,
        "Payer": "AWM9vmGpAhFyiXxg8r5Cx4H3mS2zrtSkUF",
        "TxType": 209,
        "Payload": {
            "Code": "00c66b149fdd13f41303beb7771ddd0aad6b2d815dcd62916a7cc81400000000000000000000000000000000000000016a7cc8149fdd13f41303beb7771ddd0aad6b2d815dcd62916a7cc8085da07645000000006a7cc86c0c7472616e7366657246726f6d1400000000000000000000000000000000000000020068164f6e746f6c6f67792e4e61746976652e496e766f6b65"
        },
        "Attributes": [],
        "Sigs": [
            {
                "PubKeys": [
                    "03e9ac636107c8d5a22e87bf6ae76a5e7a1394930972db72e0c3bebf54e8210a37"
                ],
                "M": 1,
                "SigData": [
                    "01dfcf5328a6587b2e2b30d6fae73bc18343ce7e5db2c00b3c92415a7274cfb1367d74604121dfd2eb8aef95b1a5e688bdde5633f1bde0fe85881db55ea2fd112d"
                ]
            }
        ],
        "Hash": "5623dbd283a99ff1cd78068cba474a22bed97fceba4a56a9d38ab0fbc178c4ab",
        "Height": 175888
    },
    "Version": "1.0.0"
}
```

### 10. sendrawtransaction

向ontology网络发送交易。

如果 preExec=1，则交易为预执行。


#### Request Example:

```
{
    "Action":"sendrawtransaction",
    "Version":"1.0.0",
    "Id":12345, //optional
    "PreExec": 0,
    "Data":"80000001195876cb34364dc38b730077156c6bc3a7fc570044a66fbfeeea56f71327e8ab0000029b7cffdaa674beae0f930ebe6085af9093e5fe56b34a5c220ccdcf6efc336fc500c65eaf440000000f9a23e06f74cf86b8827a9108ec2e0f89ad956c9b7cffdaa674beae0f930ebe6085af9093e5fe56b34a5c220ccdcf6efc336fc50092e14b5e00000030aab52ad93f6ce17ca07fa88fc191828c58cb71014140915467ecd359684b2dc358024ca750609591aa731a0b309c7fb3cab5cd0836ad3992aa0a24da431f43b68883ea5651d548feb6bd3c8e16376e6e426f91f84c58232103322f35c7819267e721335948d385fae5be66e7ba8c748ac15467dcca0693692dac"
}
```
可以使用ontology-go-sdk生成十六进制数据，参考这个[例子](rpc_api_CN.md#8-sendrawtransaction)

#### Response Example:
```
{
    "Action": "sendrawtransaction",
    "Desc": "SUCCESS",
    "Error": 0,
    "Result": "22471ab3f4b4307a99f00c9a717dbf8b26f5bf63bf47f9c560477da8181de777",
    "Version": "1.0.0"
}
```
> Result: 交易哈希

### 11. getstorage

通过合约地址哈希和键得到对应的值。

合约地址哈希的生成方式如下：

```
    addr := types.AddressFromVmCode([]byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x04})
    fmt.Println(addr.ToHexString())
```

#### Request Example
```
{
    "Action": "getstorage",
    "Version": "1.0.0",
    "Id":12345, //optional
    "Hash": "0144587c1094f6929ed7362d6328cffff4fb4da2",
    "Key" : "4587c1094f6"
}
```
#### Response Example
```
{
    "Action": "getstorage",
    "Desc": "SUCCESS",
    "Error": 0,
    "Result": "58d15e17628000",
    "Version": "1.0.0"
}
```
> 注意: 返回的值和传入的key参数均是十六进制。

### 12. getbalance

得到该地址的账户的余额。


#### Request Example
```
{
    "Action": "getbalance",
    "Version": "1.0.0",
    "Id":12345, //optional
    "Addr": "TA63xZXqdPLtDeznWQ6Ns4UsbqprLrrLJk"
}
```

#### Response Example
```
{
    "Action": "getbalance",
    "Desc": "SUCCESS",
    "Error": 0,
    "Result": {
        "ont": "2500",
        "ong": "0"
    },
    "Version": "1.0.0"
}
```
### 13. getcontract

根据合约地址哈希得到合约信息。


#### Request Example:

```
{
    "Action": "getcontract",
    "Version": "1.0.0",
    "Id":12345, //optional
    "Hash": "0100000000000000000000000000000000000000"
}
```

#### Response Example:

```
{
    "Action": "getcontract",
    "Desc": "SUCCESS",
    "Error": 0,
    "Version": "1.0.0",
    "Result": {
        "Code": "0000000000000000000000000000000000000001",
        "NeedStorage": true,
        "Name": "ONT",
        "CodeVersion": "1.0",
        "Author": "Ontology Team",
        "Email": "contact@ont.io",
        "Description": "Ontology Network ONT Token"
    }
}
```

#### 14. getsmartcodeeventbyheight

得到该高度区块上的智能合约执行结果。


#### Example usage:

```
{
    "Action": "getsmartcodeeventbyheight",
    "Version": "1.0.0",
    "Id":12345, //optional
    "Height": 100
}
```

#### Response Example
```
{
    "Action": "getsmartcodeeventbyheight",
    "Desc": "SUCCESS",
    "Error": 0,
    "Result": [
               {
                    "TxHash": "7e8c19fdd4f9ba67f95659833e336eac37116f74ea8bf7be4541ada05b13503e",
                    "State": 1,
                    "GasConsumed": 0,
                    "Notify": [
                        {
                            "ContractAddress": "0200000000000000000000000000000000000000",
                            "States": [
                                "transfer",
                                "AFmseVrdL9f9oyCzZefL9tG6UbvhPbdYzM",
                                "AFmseVrdL9f9oyCzZefL9tG6UbvhUMqNMV",
                                1000000000000000000
                            ]
                        }
                    ]
                },
                {
                    "TxHash": "fc82cd363271729367098fbabcfd0c02cf6ded1e535700d04658b596d53cf07d",
                    "State": 1,
                    "GasConsumed": 0,
                    "Notify": [
                        {
                            "ContractAddress": "0200000000000000000000000000000000000000",
                            "States": [
                                "transfer",
                                "AFmseVrdL9f9oyCzZefL9tG6UbvhPbdYzM",
                                "AFmseVrdL9f9oyCzZefL9tG6UbvhUMqNMV",
                                1000000000000000000
                            ]
                        }
                    ]
                }
    ],
    "Version": "1.0.0"
}
```
> 注意: 返回的结果是交易简略信息的集合，并不是完整的交易信息。

### 15. getsmartcodeeventbyhash

通过交易哈希得到该交易的执行结果。

#### Request Example:
```
{
    "Action": "getsmartcodeeventbyhash",
    "Version": "1.0.0",
    "Id":12345, //optional
    "Hash": "20046da68ef6a91f6959caa798a5ac7660cc80cf4098921bc63604d93208a8ac"
}
```
#### Response Example:
```
{
    "Action": "getsmartcodeeventbyhash",
    "Desc": "SUCCESS",
    "Error": 0,
    "Version": "1.0.0",
    "Result": {
             "TxHash": "20046da68ef6a91f6959caa798a5ac7660cc80cf4098921bc63604d93208a8ac",
             "State": 1,
             "GasConsumed": 0,
             "Notify": [
                    {
                      "ContractAddress": "ff00000000000000000000000000000000000001",
                      "States": [
                            "transfer",
                            "A9yD14Nj9j7xAB4dbGeiX9h8unkKHxuWwb",
                            "AA4WVfUB1ipHL8s3PRSYgeV1HhAU3KcKTq",
                            1000000000
                         ]
                     }
              ]
    }
}
```
### 16. getblockheightbytxhash

通过交易哈希得到该交易落账的区块高度。

#### Request Example:
```
{
    "Action": "getblockheightbytxhash",
    "Version": "1.0.0",
    "Id":12345, //optional
    "Hash": "3e23cf222a47739d4141255da617cd42925a12638ac19cadcc85501f907972c8"
}
```
#### Response Example
```
{
    "Action": "getblockheightbytxhash",
    "Desc": "SUCCESS",
    "Error": 0,
    "Result": 100,
    "Version": "1.0.0"
}
```


### 17. getmerkleproof

通过交易哈希得到该交易的merkle证明。

#### Request Example:
```
{
    "Action": "getmerkleproof",
    "Version": "1.0.0",
    "Id":12345, //optional
    "Hash": "0087217323d87284d21c3539f216dd030bf9da480372456d1fa02eec74c3226d"
}

```
#### Response Example
```
{
    "Action": "getmerkleproof",
    "Desc": "SUCCESS",
    "Error": 0,
    "Result": {
        "Type": "MerkleProof",
        "TransactionsRoot": "fe3a4ee8a44e3e588de55de1b8fe08f08b6184d9c062cf7316fb9481eb57b9e6",
        "BlockHeight": 600,
        "CurBlockRoot": "57476eba688531dec8555cb712835c7eda48a478431a2cfd3372aeee5298e711",
        "CurBlockHeight": 6478,
        "TargetHashes": [
            "270cd10ea235cc18cba83a070fdf18ae576983b6b9a7bb9a3fec540b3786c85c",
            "24e4697f9dd6cb944d0736bd3e11b64f64edec94fb599e25d4e5461d54174f0e",
            "9a47ab04acf6bba7bb97b83eddeb0db20e11c0627b8079b40b60031d5bd63154",
            "d1b513810b9b983014c9f8b7084b8ea8744eca8e7c942586c2b7c63f910363ca",
            "54e88360efedcf5dbbc486ea0267724a98b027b3ba780617e32569bb3fbe56e8",
            "e0c5ebca3ca191617d42e11db64778b047cd9a520538efd95d5a688cbba0c8d5",
            "52bfb23b6456cac4e5e7143287e1518dd923c5b5d32d0bfe8d825dc8195ea62b",
            "86d6be166ae1a53c052adc40b9b66c4f95f5e3b6ecc88afaea3750e1cbe98276",
            "5588530cfc4d92e979717f8ae399ac4553a76e7537a981e8eaf078d60f1d39a6",
            "3f15bec38bcf054e4f32efe475a09d3e80c2e90d3345a1428aaa262606f13267",
            "f238ed8ceb1c10a08f7eaa390cdde44ed7d160abbde4702028407b55671e7aa8",
            "b4813f1f27c0457726b58f8bf20bee70c100a4d5c5f1805e53dcd20f38479615",
            "83893713ea8ace9214b28af854b75671c8aaa62bb74b0d43ad6fb83e3dee42db"
        ]
    },
    "Version": "1.0.0"
}
```

### 18. getsessioncount

得到会话数量。

#### Request Example:
```
{
    "Action": "getsessioncount",
    "Version": "1.0.0",
    "Id":12345, //optional
}
```
#### Response Example
```
{
    "Action": "getsessioncount",
    "Desc": "SUCCESS",
    "Error": 0,
    "Result": 10,
    "Version": "1.0.0"
}
```

### 19. getgasprice

得到gas的价格。

#### Request Example:
```
{
    "Action": "getgasprice",
    "Version": "1.0.0",
    "Id":12345, //optional
}
```
#### Response Example
```
{
    "Action": "getgasprice",
    "Desc": "SUCCESS",
    "Error": 0,
    "Result": {
         "gasprice": 0,
         "height": 1
     },
    "Version": "1.0.0"
}
```

### 20. getallowance

得到允许从from账户转出到to账户的额度。

#### Request Example:
```
{
    "Action": "getallowance",
    "Id":12345, //optional
    "Asset": "ont",
    "From" :  "A9yD14Nj9j7xAB4dbGeiX9h8unkKHxuWwb",
    "To"   :  "AA4WVfUB1ipHL8s3PRSYgeV1HhAU3KcKTq",
    "Version": "1.0.0"
}
```
#### Response Example
```
{
    "Action": "getallowance",
    "Desc": "SUCCESS",
    "Error": 0,
    "Result": "10",
    "Version": "1.0.0"
}
```

### 21. getunboundong

得到该账户未提取的ong数量。

#### Request Example:
```
{
    "Action": "getunboundong",
    "Id":12345, //optional
    "Addr": "ANH5bHrrt111XwNEnuPZj6u95Dd6u7G4D6",
    "Version": "1.0.0"
}
```
#### Response Example
```
{
    "Action": "getunboundong",
    "Desc": "SUCCESS",
    "Error": 0,
    "Result": "204957950400000",
    "Version": "1.0.0"
}
```

### 22. getmempooltxstate

通过交易哈希得到内存中该交易的状态。

#### Request Example:
```
{
    "Action": "getmempooltxstate",
    "Id":12345, //optional
    "Hash": "0b437771a42d18d292741c5d4f1300a135fa6e65b0594e39dc299e7f8279221a",
    "Version": "1.0.0"
}
```
#### Response Example
```
{
    "Action": "getmempooltxstate",
    "Desc": "SUCCESS",
    "Error": 0,
    "Version": "1.0.0",
    "Result": {
              	"State": [{
              		"Type": 1,
              		"Height": 342,
              		"ErrCode": 0
              	}, {
              		"Type": 0,
              		"Height": 0,
              		"ErrCode": 0
              	}]
    }
}
```

### 23. getmempooltxcount

得到内存中的交易的数量。

#### Request Example:
```
{
    "Action": "getmempooltxcount",
    "Id":12345, //optional
    "Version": "1.0.0"
}
```
#### Response Example
```
{
    "Action": "getmempooltxcount",
    "Desc": "SUCCESS",
    "Error": 0,
    "Version": "1.0.0",
    "Result": [100,50]
}
```


### 24. getversion

得到版本信息。

#### Request Example:
```
{
    "Action": "getversion",
    "Id":12345, //optional
    "Version": "1.0.0"
}
```
#### Response Example
```
{
    "Action": "getversion",
    "Desc": "SUCCESS",
    "Error": 0,
    "Version": "1.0.0",
    "Result": "0.9"
}
```

### 25. getnetworkid

获取 network id

#### Request Example:
```
{
    "Action": "getnetworkid",
    "Id":12345, //optional
    "Version": "1.0.0"
}
```
#### Response Example
```
{
    "Action": "getnetworkid",
    "Desc": "SUCCESS",
    "Error": 0,
    "Version": "1.0.0",
    "Result": 1
}
```

### 26. getgrantong

获取 grant ong

#### Request Example:
```
{
    "Action": "getgrantong",
    "Id":12345, //optional
    "Addr":"AKDFapcoUhewN9Kaj6XhHusurfHzUiZqUA",
    "Version": "1.0.0"
}
```
#### Response Example
```
{
    "Action": "getgrantong",
    "Desc": "SUCCESS",
    "Error": 0,
    "Version": "1.0.0",
    "Result": 4995625
}
```

## 错误代码

| Field | Type | Description |
| :--- | :--- | :--- |
| 0 | int64 | SUCCESS |
| 41001 | int64 | SESSION\_EXPIRED: 无效或超时的会话 |
| 41002 | int64 | SERVICE\_CEILING: 达到服务上限 |
| 41003 | int64 | ILLEGAL\_DATAFORMAT: 不合法的数据格式 |
| 41004 | int64 | INVALID\_VERSION: 无效的版本号 |
| 42001 | int64 | INVALID\_METHOD: 无效的方法 |
| 42002 | int64 | INVALID\_PARAMS: 无效的参数 |
| 43001 | int64 | INVALID\_TRANSACTION: 无效的交易 |
| 43002 | int64 | INVALID\_ASSET: 无效的资源 |
| 43003 | int64 | INVALID\_BLOCK: 无效的区块 |
| 44001 | int64 | UNKNOWN\_TRANSACTION: 未知的交易 |
| 44002 | int64 | UNKNOWN\_ASSET: 未知的资源 |
| 44003 | int64 | UNKNOWN\_BLOCK: 未知的区块 |
| 45001 | int64 | INTERNAL\_ERROR: 内部错误 |
| 47001 | int64 | SMARTCODE\_ERROR: 智能合约执行错误 |
