# Ontology Restful API

[English](restful_api.md)|中文

* [介绍](#介绍)
* [Restful接口列表](#restful接口列表)
* [错误代码](#错误代码)

## 介绍

本文档是Ontology的REST接口文档，详细定义了各个接口所需的参数与返回值。

### 响应参数定义

| Field | Type | Description |
| :--- | :--- | :--- |
| Action | string | 响应动作名称 |
| Desc | string | 响应结果描述 |
| Error | int64 | 错误代码 |
| Result | object | 执行结果 |
| Version | string | 版本号 |

## Restful接口列表

| Method | URL | Description |
| :---| :---| :---|
| [get_conn_count](#1-get_conn_count) | GET /api/v1/node/connectioncount | 得到当前连接的节点数量 |
| [get_blk_txs_by_height](#2-get_blk_txs_by_height) | GET /api/v1/block/transactions/height/:height | 得到该高度的区块的所有交易哈希 |
| [get_blk_by_height](#3-get_blk_by_height) | GET /api/v1/block/details/height/:height?raw=0 | 得到该高度的区块信息 |
| [get_blk_by_hash](#4-get_blk_by_hash) | GET /api/v1/block/details/hash/:hash?raw=1 | 通过区块哈希得到区块信息 |
| [get_blk_height](#5-get_blk_height) | GET /api/v1/block/height | 得到当前网络上的区块高度 |
| [get_blk_hash](#6-get_blk_hash) | GET /api/v1/block/hash/:height | 得到该高度的区块哈希 |
| [get_tx](#7-get_tx) | GET /api/v1/transaction/:hash | 通过交易哈希得到该交易的信息 |
| [get_storage](#8-get_storage) | GET /api/v1/storage/:hash/:key| 通过合约地址哈希和键得到对应的值 |
| [get_balance](#9-get_balance) | GET /api/v1/balance/:addr | 得到该地址的账户的余额 |
| [get_contract_state](#10-get_contract_state) | GET /api/v1/contract/:hash | 得到合约状态 |
| [get_sc_event_by_height](#11-get_sc_event_by_height) | GET /api/v1/smartcode/event/transactions/:height | 得到该高度区块上的智能合约执行结果 |
| [get_smtcode_evts](#12-get_smtcode_evts) | GET /api/v1/smartcode/event/txhash/:hash | 通过交易哈希得到该交易的执行结果 |
| [get_blk_hgt_by_txhash](#13-get_blk_hgt_by_txhash) | GET /api/v1/block/height/txhash/:hash | 通过交易哈希得到该交易落账的区块高度 |
| [get_merkle_proof](#14-get_merkle_proof) | GET /api/v1/merkleproof/:hash| 通过交易哈希得到该交易的merkle证明 |
| [get_gasprice](#15-get_gasprice) | GET /api/v1/gasprice| 得到gas的价格 |
| [get_allowance](#16-get_allowance) | GET /api/v1/allowance/:asset/:from/:to | 返回允许从from账户转出到to账户的额度 |
| [get_unboundong](#17-get_unboundong) | GET /api/v1/unboundong/:addr | 返回该账户未提取的ong数量 |
| [get_mempooltxcount](#18-get_mempooltxcount) | GET /api/v1/mempool/txcount | 得到内存中的交易的数量 |
| [get_mempooltxstate](#19-get_mempooltxstate) | GET /api/v1/mempool/txstate/:hash | 通过交易哈希得到内存中该交易的状态 |
| [get_version](#20-get_version) |  GET /api/v1/version | 得到版本信息 |
| [post_raw_tx](#21-post_raw_tx) | post /api/v1/transaction?preExec=0 | 向ontology网络发送交易 |
| [get_networkid](#22-get_networkid) |  GET /api/v1/networkid | 得到network id |
| [get_grantong](#23-get_grantong) |  GET /api/v1/grantong/:addr | 得到grant ong |

### 1 get_conn_count

得到当前连接的节点数量。


GET

```
/api/v1/node/connectioncount
```

#### Request Example:

```
curl -i http://server:port/api/v1/node/connectioncount
```

#### Response Example:

```
{
    "Action": "getconnectioncount",
    "Desc": "SUCCESS",
    "Error": 0,
    "Result": 0,
    "Version": "1.0.0"
}
```
### 2 get_blk_txs_by_height

得到该高度的区块的所有交易哈希。

GET

```
/api/v1/block/transactions/height/:height
```

#### Request Example:

```
curl -i http://server:port/api/v1/block/transactions/height/100
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
### 3 get_blk_by_height

得到该高度的详细的区块信息。

raw：可选参数，默认值为零，不设置时为默认值。当值为1时，接口返回区块序列化后的信息，该信息以十六进制字符串表示。如果要得到区块的具体信息，需要调用
 SDK中的方法对该字符串进行反序列化。当值为0时，将以json格式返回对应区块的详细信息。

GET

```
/api/v1/block/details/height/:height?raw=1
```

#### Request Example:

```
curl -i http://server:port/api/v1/block/details/height/22
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
### 4 get_blk_by_hash

通过区块哈希得到区块信息。

raw：可选参数，默认值为零，不设置时为默认值。当值为1时，接口返回区块序列化后的信息，该信息以十六进制字符串表示。如果要得到区块的具体信息，需要调用
 SDK中的方法对该字符串进行反序列化。当值为0时，将以json格式返回对应区块的详细信息。

GET

```
/api/v1/block/details/hash/:hash?raw=0
```

#### Request Example:

```
curl -i http://server:port/api/v1/block/details/hash/ea5e5219d2f1591f4feef89885c3f38c83d3a3474a5622cf8cd3de1b93849603
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

### 5 get_blk_height

得到当前网络上的区块高度。


GET

```
/api/v1/block/height
```

#### Request Example:

```
curl -i http://server:port/api/v1/block/height
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

### 6 get_blk_hash

得到该高度的区块哈希。

GET

```
/api/v1/block/hash/:height
```

#### Request Example:

```
curl -i http://server:port/api/v1/block/hash/100
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

### 7 get_tx

通过交易哈希得到该交易的信息。

raw：可选参数，默认值为零，不设置时为默认值。当值为1时，接口返回交易序列化后的信息，该信息以十六进制字符串表示。如果要得到交易的具体信息，需要调用
 SDK中的方法对该字符串进行反序列化。当值为0时，将以json格式返回对应交易的详细信息。

GET

```
/api/v1/transaction/:hash?raw=0
```

#### Request Example:

```
curl -i http://server:port/api/v1/transaction/5623dbd283a99ff1cd78068cba474a22bed97fceba4a56a9d38ab0fbc178c4ab
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

### 8 get_storage

通过合约地址哈希和键得到对应的值。

合约地址哈希的生成方式如下：

```
    addr := types.AddressFromVmCode([]byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x04})
    fmt.Println(addr.ToHexString())
```


GET
```
/api/v1/storage/:hash/:key
```
#### Request Example
```
curl -i http://localhost:20334/api/v1/storage/ff00000000000000000000000000000000000001/0144587c1094f6929ed7362d6328cffff4fb4da2
```
#### Response
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

### 9 get_balance

得到该地址的账户的余额。

GET
```
/api/v1/balance/:addr
```
> addr: Base58编码后的账户地址

#### Request Example
```
curl -i http://localhost:20334/api/v1/balance/TA5uYzLU2vBvvfCMxyV2sdzc9kPqJzGZWq
```

#### Response
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
### 10 get_contract_state

根据合约地址哈希得到合约信息。

GET

```
/api/v1/contract/:hash
```

#### Request Example:

```
curl -i http://server:port/api/v1/contract/0100000000000000000000000000000000000000
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

#### 11 get_sc_event_by_height

得到该高度区块上的智能合约执行结果。

GET

```
/api/v1/smartcode/event/transactions/:height
```

#### Example usage:

```
curl -i http://localhost:20334/api/v1/smartcode/event/transactions/900
```

#### response
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

### 12 get_smtcode_evts

通过交易哈希得到该交易的执行结果。

GET
```
/api/v1/smartcode/event/txhash/:hash
```
#### Request Example:
```
curl -i http://localhost:20334/api/v1/smartcode/event/txhash/20046da68ef6a91f6959caa798a5ac7660cc80cf4098921bc63604d93208a8ac
```
#### Response:
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
                            "T9yD14Nj9j7xAB4dbGeiX9h8unkKHxuWwb",
                            "TA4WVfUB1ipHL8s3PRSYgeV1HhAU3KcKTq",
                            1000000000
                         ]
                     }
              ]
    }
}
```
### 13 get_blk_hgt_by_txhash

通过交易哈希得到该交易落账的区块高度。

GET
```
/api/v1/block/height/txhash/:hash
```
#### Request Example:
```
curl -i http://localhost:20334/api/v1/block/height/txhash/3e23cf222a47739d4141255da617cd42925a12638ac19cadcc85501f907972c8
```
#### Response
```
{
    "Action": "getblockheightbytxhash",
    "Desc": "SUCCESS",
    "Error": 0,
    "Result": 0,
    "Version": "1.0.0"
}
```

### 14 get_merkle_proof

通过交易哈希得到该交易的merkle证明。

GET
```
/api/v1/merkleproof/:hash
```
#### Request Example:
```
curl -i http://localhost:20334/api/v1/merkleproof/3e23cf222a47739d4141255da617cd42925a12638ac19cadcc85501f907972c8
```
#### Response
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

### 15 get_gasprice

得到gas的价格。

GET
```
/api/v1/gasprice
```
#### Request Example:
```
curl -i http://localhost:20334/api/v1/block/height/txhash/3e23cf222a47739d4141255da617cd42925a12638ac19cadcc85501f907972c8
```
#### Response
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

### 16 get_allowance

得到允许从from账户转出到to账户的额度。

GET
```
/api/v1/allowance
```
#### Request Example:
```
curl -i http://localhost:20334/api/v1/allowance/:asset/:from/:to
```
#### Response
```
{
    "Action": "getallowance",
    "Desc": "SUCCESS",
    "Error": 0,
    "Result": "10",
    "Version": "1.0.0"
}
```

### 17 get_unboundong

得到该账户未提取的ong数量。

GET
```
/api/v1/unboundong
```
#### Request Example:
```
curl -i http://localhost:20334/api/v1/unboundong/:addr
```
#### Response
```
{
    "Action": "getunboundong",
    "Desc": "SUCCESS",
    "Error": 0,
    "Result": "204957950400000",
    "Version": "1.0.0"
}
```

### 18 get_mempooltxcount

得到内存中的交易的数量。

GET
```
/api/v1/mempool/txcount
```
#### Request Example:
```
curl -i http://localhost:20334/api/v1/mempool/txcount
```
#### Response
```
{
    "Action": "getmempooltxcount",
    "Desc": "SUCCESS",
    "Error": 0,
    "Version": "1.0.0",
    "Result": [100,50]
}
```

### 19 get_mempooltxstate

通过交易哈希得到内存中该交易的状态。

GET
```
/api/v1/mempool/txstate/:hash
```
#### Request Example:
```
curl -i http://localhost:20334/api/v1/mempool/txstate/:hash
```
#### Response
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

### 20 get_version

得到版本信息。

GET
```
/api/v1/version
```
#### Request Example:
```
curl -i http://localhost:20334/api/v1/version
```
#### Response
```
{
    "Action": "getversion",
    "Desc": "SUCCESS",
    "Error": 0,
    "Version": "1.0.0",
    "Result": "0.9"
}
```

### 21 post_raw_tx

向ontology网络发送交易。

如果 preExec=1，则交易为预执行。

POST

```
/api/v1/transaction?preExec=0
```

#### Request Example:

```
curl  -H "Content-Type: application/json"  -X POST -d '{"Action":"sendrawtransaction", "Version":"1.0.0","Data":"00d00000000080fdcf2b0138c56b6c766b00527ac46c766b51527ac46151c56c766b52527ac46c766b00c31052656749644279507..."}'  http://server:port/api/v1/transaction
```

#### Post Params:

```
{
    "Action":"sendrawtransaction",
    "Version":"1.0.0",
    "Data":"80000001195876cb34364dc38b730077156c6bc3a7fc570044a66fbfeeea56f71327e8ab0000029b7cffdaa674beae0f930ebe6085af9093e5fe56b34a5c220ccdcf6efc336fc500c65eaf440000000f9a23e06f74cf86b8827a9108ec2e0f89ad956c9b7cffdaa674beae0f930ebe6085af9093e5fe56b34a5c220ccdcf6efc336fc50092e14b5e00000030aab52ad93f6ce17ca07fa88fc191828c58cb71014140915467ecd359684b2dc358024ca750609591aa731a0b309c7fb3cab5cd0836ad3992aa0a24da431f43b68883ea5651d548feb6bd3c8e16376e6e426f91f84c58232103322f35c7819267e721335948d385fae5be66e7ba8c748ac15467dcca0693692dac"
}
```
可以使用ontology-go-sdk生成十六进制数据，参考这个[例子](rpc_api_CN.md#8-sendrawtransaction)

#### Response
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


### 22 get_networkid

获取 network id.

GET
```
/api/v1/networkid
```
#### Request Example:
```
curl -i http://localhost:20334/api/v1/networkid
```
#### Response
```
{
    "Action": "getnetworkid",
    "Desc": "SUCCESS",
    "Error": 0,
    "Version": "1.0.0",
    "Result": 1
}
```

### 23 get_grantong

获取 grant ong.

GET
```
/api/v1/grantong/:addr
```
#### Request Example:
```
curl -i http://localhost:20334/api/v1/grantong/AKDFapcoUhewN9Kaj6XhHusurfHzUiZqUA
```
#### Response
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
