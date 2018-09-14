# ONT Rpc Api

[English](rpc_api.md)|中文

* [介绍](#介绍)
* [RPC接口列表](#rpc接口列表)
* [错误代码](#错误代码)

## 介绍

本文档是Ontology的RPC接口文档，详细定义了每种接口的参数与返回值。

以下是一些接口中用到的字段的定义：

#### 请求参数定义:

| 字段 | 类型 | 定义 |
| :---| :---| :---|
| jsonrpc | string | jsonrpc版本号 |
| method | string | 方法名 |
| params | string | 方法要求的参数 |
| id | int | 任意值 |

#### 相应参数定义:

| 字段 | 类型 | 定义 |
| :---| :---| :---|
| desc| string | 请求结果描述 |
| error | int64 | 错误代码 |
| jsonrpc | string | jsonrpc版本号 |
| id | int | 任意值 |
| result | object | RPC执行结果 |

>注意: 不同的请求类型会返回不同类型的Result。

#### 区块字段定义：

| 字段 | 类型 | 定义 |
| :--- | :--- | :--- |
| Header | *Header |  |
| Transactions | []*Transaction ||
| hash | *Uint256 | |

#### 区块头字段定义

| 字段 | 类型 | 定义 |
| :--- | :--- | :--- |
| Version | uint32 | 版本号 |
| PrevBlockHash | Uint256 | 前一个区块的哈希 |
| TransactionsRoot | Uint256 | 该区块中所有交易的Merkle树树根 |
| BlockRoot | Uint256 | 区块根 |
| Timestamp | int | 区块时间戳，unix时间格式 |
| Height | int | 区块高度 |
| ConsensusData | uint64 |  |
| NextBookkeeper | Address | 下一个记账人的地址 |
| Bookkeepers | []*crypto.PubKey ||
| SigData | [][]byte ||
| Hash | Uint256 | 区块哈希 |

#### 交易字段定义

| 字段 | 类型 | 定义 |
| :--- | :--- | :--- |
| Version| byte | 版本号 |
| TxType | TransactionType | 交易类型 |
| Payload | Payload | 载荷，具体执行的交易数据 |
| Nonce | uint32 | 随机值，可以设置为时间戳 |
| Attributes | []*TxAttribute |  |
| Fee | []*Fee | 交易费用  |
| NetworkFee | Fixed64 | 网络费用 |
| Sigs | []*Sig | 签名数据 |
| Hash | *Uint256 | 交易哈希 |

## RPC接口列表

| Method | Parameters | Description | Note |
| :---| :---| :---| :---|
| [getbestblockhash](#1-getbestblockhash) |  | 得到主链上的最高区块的哈希 |  |
| [getblock](#2-getblock) | height or blockhash,[verbose] | 通过区块哈希或高度得到区块 | verbose为可选参数，默认值为0，可选值为1 |
| [getblockcount](#3-getblockcount) |  | 得到区块的数量 |  |
| [getblockhash](#4-getblockhash) | height | 得到对应高度的区块的哈希 |  |
| [getconnectioncount](#5-getconnectioncount)|  | 得到当前网络上连接的节点数 |  |
| [getrawtransaction](#6-getrawtransaction) | transactionhash | 通过交易哈希得到交易详情 |  |
| [sendrawtransaction](#7-sendrawtransaction) | hex,preExec | 向网络中发送交易 | 发送的数据为签过名的交易序列化后的十六进制字符串 |
| [getstorage](#8-getstorage) | script_hash, key |根据合约地址和存储的键，得到对应的值 |  |
| [getversion](#9-getversion) |  | 得到运行的ontology版本 |  |
| [getcontractstate](#10-getcontractstate) | script_hash,[verbose] | 根据合约地址，得到合约信息 |  |
| [getmempooltxcount](#11-getmempooltxcount) |         | 查询内存中的交易的数量 |  |
| [getmempooltxstate](#12-getmempooltxstate) | tx_hash | 查询内存中的交易的状态 |  |
| [getsmartcodeevent](#13-getsmartcodeevent) |  | 得到智能合约执行的结果 |  |
| [getblockheightbytxhash](#14-getblockheightbytxhash) | tx_hash | 得到该交易哈希所落账的区块的高度 |  |
| [getbalance](#15-getbalance) | address | 返回base58地址的余额 |  |
| [getmerkleproof](#16-getmerkleproof) | tx_hash | 返回merkle证明 |  |
| [getgasprice](#17-getgasprice) |  | 返回gas的价格 |  |
| [getallowance](#18-getallowance) | asset, from, to | 返回允许从from转出到to账户的额度 |  |
| [getunboundong](#19-getunboundong) | address | 返回该账户未提取的ong |  |
| [getblocktxsbyheight](#20-getblocktxsbyheight) | height | 返回该高度对应的区块落账的交易的哈希 |  |
| [getnetworkid](#21-getnetworkid) |  | 获取 network id |  |
| [getgrantong](#22-getgrantong) |  | 获取 grant ong |  |

### 1. getbestblockhash

得到主链上的最高区块的哈希。

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

### 2. getblock

通过区块哈希或高度得到区块。

#### 参数定义
Hash/height: 区块哈希/高度

Verbose: 可选参数，默认值为零，不设置时为默认值。当值为0时，接口返回区块序列化后的信息，该信息以十六进制字符串表示。如果要得到区块的具体信息，需要调用
SDK中的方法对该字符串进行反序列化。当值为1时，将以json格式返回对应区块的详细信息。

#### Example

Request:

```
{
  "jsonrpc": "2.0",
  "method": "getblock",
  "params": ["773dd2dae4a9c9275290f89b56e67d7363ea4826dfd4fc13cc01cf73a44b0d0e"],
  "id": 1
}
```

or

```
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
    "jsonrpc": "2.0",
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

得到主链上的区块总量。

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

>Result: 主链高度。

#### 4. getblockhash

返回对应高度的区块哈希。

#### 参数定义

Index: 区块高度

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

得到当前网络上连接的节点数。

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


#### 6. getrawtransaction

通过交易哈希得到交易详情。

#### 参数定义

txid: 交易哈希

Verbose: 可选参数，默认值为零，不设置时为默认值。当值为0时，接口返回交易序列化后的信息，该信息以十六进制字符串表示。如果要得到交易的具体信息，需要调用
SDK中的方法对该字符串进行反序列化。当值为1时，将以json格式返回对应交易的详细信息。

#### Example

When verbose is nil or verbose = 0:

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

When verbose = 1:

Request:

```
{
  "jsonrpc": "2.0",
  "method": "getrawtransaction",
  "params": ["5623dbd283a99ff1cd78068cba474a22bed97fceba4a56a9d38ab0fbc178c4ab", 1],
  "id": 1
}
```
Response:

```
{
    "desc": "SUCCESS",
    "error": 0,
    "id": 1,
    "jsonrpc": "2.0",
    "result": {
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
    }
}
```



#### 7. sendrawtransaction

发送交易。

#### 参数定义

Hex: 签名后的交易序列化成的十六进数据。可以参考ontology-go-sdk/rpc.go中的NewNativeInvokeTransaction方法生成。

PreExec : 值设置为1则表示此交易为预执行。

如何生成交易参数（Hex）？

```
    // 得到SDK的实例
    sdk := goSdk.NewOntologySdk()
    rpcClient := sdk.Rpc
    // 生成native合约调用交易; 如果想调用NEO VM合约，可以使用NewNeoVMSInvokeTransaction方法
    // cversion 为合约的版本, method 是要调用的合约方法名, params 是该方法需要的参数
    // 例如：
    // NewNativeInvokeTransaction(0, 200000, byte(0),utils.ParamContractAddress,
    //      "getGlobalParam", []interface{}{global_params.ParamNameList{"gasPrice"}})
    tx, err := rpcClient.NewNativeInvokeTransaction(gasPrice, gasLimit, cversion, contractAddress, method, params)
    if err != nil {
    	return common.UINT256_EMPTY, err
    }
    // 对交易签名，signer为交易的发送者
    err = rpcClient.SignToTransaction(tx, signer)
    if err != nil {
        return common.UINT256_EMPTY, err
    }

    txbf := new(bytes.Buffer)
    err = tx.Serialize(txbf);
    hexCode = common.ToHexString(txbf.Bytes())
```

相关的结构体
```
type Transaction struct {
	Version  byte
	TxType   TransactionType
	Nonce    uint32
	GasPrice uint64
	GasLimit uint64
	Payer    common.Address
	Payload  Payload
	attributes byte
	Sigs       []*Sig

	hash *common.Uint256
}

type Sig struct {
	SigData [][]byte
	PubKeys []keypair.PublicKey
	M       uint16
}
```

#### Example

Request:

```
{
  "jsonrpc": "2.0",
  "method": "sendrawtransaction",
  "params": ["80000001195876cb34364dc38b730077156c6bc3a7fc570044a66fbfeeea56f71327e8ab0000029b7cffdaa674beae0f930ebe6085af9093e5fe56b34a5c220ccdcf6efc336fc500c65eaf440000000f9a23e06f74cf86b8827a9108ec2e0f89ad956c9b7cffdaa674beae0f930ebe6085af9093e5fe56b34a5c220ccdcf6efc336fc50092e14b5e00000030aab52ad93f6ce17ca07fa88fc191828c58cb71014140915467ecd359684b2dc358024ca750609591aa731a0b309c7fb3cab5cd0836ad3992aa0a24da431f43b68883ea5651d548feb6bd3c8e16376e6e426f91f84c58232103322f35c7819267e721335948d385fae5be66e7ba8c748ac15467dcca0693692dac",0],
  "id": 1
}
```

Reponse

```
{
    "desc": "SUCCESS",
    "error": 0,
    "id": 1,
    "jsonrpc": "2.0",
    "result": "498db60e96828581eff991c58fa46abbfd97d2f4a4f9915a11f85c54f2a2fedf"
}
```

> 注意：返回的结果是交易哈希

#### 8. getstorage

根据合约地址和存储的键，得到对应的值。

#### 参数定义

script\_hash: 合约地址哈希，通过以下方法生成：

```
	addr := types.AddressFromVmCode([]byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
	    0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x04})
	fmt.Println(addr.ToHexString())
```

Key: 存储的条目的键，要求转化成十六进制字符串

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
> 返回结果为十六进制字符串

#### 9. getversion

得到运行的ontology版本。

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
  "result": "v0.9.2-1-g231e"
}
```

#### 10. getcontractstate

根据合约地址，得到对应的合约信息。

#### 参数定义

script\_hash: 合约地址哈希。

verbose: 可选参数，默认值为零，不设置时为默认值。当值为0时，接口返回合约序列化后的信息，该信息以十六进制字符串表示。如果要得到交易的具体信息，需要调用
SDK中的方法对该字符串进行反序列化。当值为1时，将以json格式返回对应合约的详细信息。

#### Example

Request:

```
{
  "jsonrpc": "2.0",
  "method": "getcontractstate",
  "params": ["0100000000000000000000000000000000000000",1],
  "id": 1
}
```

Response:

```
{
    "desc": "SUCCESS",
    "error": 0,
    "id": 1,
    "jsonrpc": "2.0",
    "result": {
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

#### 11. getmempooltxcount

查询内存中的交易的数量。

#### Example

Request:

```
{
  "jsonrpc": "2.0",
  "method": "getmempooltxcount",
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
    "result": [100,50]
}
```

#### 12. getmempooltxstate

查询内存中的交易的状态

#### 参数定义

tx\_hash: 交易哈希。

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

#### 13. getsmartcodeevent

得到智能合约执行的结果。

#### 参数定义

blockheight: 区块高度
或者
txHash: 交易哈希

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
  "result": [
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
        }
  ]
}
```

or

```
{
    "desc": "SUCCESS",
    "error": 0,
    "id": 1,
    "jsonrpc": "2.0",
    "result": {
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

> 注意： 如果参数是区块高度，则返回执行结果的集合；如果是交易哈希，则返回该交易对应的结果。

#### 14. getblockheightbytxhash

得到该交易哈希所落账的区块的高度。

#### 参数定义

txhash: 交易哈希

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
    "jsonrpc": "2.0",
    "result": 10
}
```

#### 15. getbalance

返回base58地址的余额

#### 参数定义

address: base58地址

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
   "jsonrpc":"2.0",
   "result":{
        "ont": "2500",
        "ong": "0"
       }
}
```

#### 16. getmerkleproof

返回对应交易的merkle证明

#### 参数定义

hash: 交易哈希

#### Example

Request:

```
{
  "jsonrpc": "2.0",
  "method": "getmerkleproof",
  "params": ["0087217323d87284d21c3539f216dd030bf9da480372456d1fa02eec74c3226d"],
  "id": 1
}
```

Response:

```
{
   "desc":"SUCCESS",
   "error":0,
   "id":1,
   "jsonrpc":"2.0",
   "result":{
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
   }
}
```

#### 17. getgasprice

返回gas价格


#### Example

Request:

```
{
  "jsonrpc": "2.0",
  "method": "getgasprice",
  "params": [],
  "id": 1
}
```

Response:

```
{
   "desc":"SUCCESS",
   "error":0,
   "id":1,
   "jsonrpc":"2.0",
   "result":{
        "gasprice": 0,
        "height": 1
       }
}
```

#### 18. getallowance

返回允许从from转出到to账户的额度

#### 参数定义

asset: "ont"或者"ong"

from: 转出账户base58地址

to: 转入账户base58地址

#### Example

Request:

```
{
  "jsonrpc": "2.0",
  "method": "getallowance",
  "params": ["ont","from address","to address"],
  "id": 1
}
```

Response:

```
{
   "desc":"SUCCESS",
   "error":0,
   "id":1,
   "jsonrpc":"2.0",
   "result": "10"
}
```

#### 19. getunboundong

返回可以提取的ong。

#### 参数定义

address：提取ong的账户地址

#### Example

Request:

```
{
  "jsonrpc": "2.0",
  "method": "getunboundong",
  "params": ["address"],
  "id": 1
}
```

Response:

```
{
   "desc":"SUCCESS",
   "error":0,
   "id":1,
   "jsonrpc":"2.0",
   "result": "204957950400000"
}
```

#### 20. getblocktxsbyheight

返回该高度对应的区块落账的所有交易的哈希

#### 参数定义

height： 区块高度

#### Example

Request:

```
{
  "jsonrpc": "2.0",
  "method": "getblocktxsbyheight",
  "params": [100],
  "id": 1
}
```

Response:

```
{
   "desc":"SUCCESS",
   "error":0,
   "id":1,
   "jsonrpc":"2.0",
   "result": {
        "Hash": "ea5e5219d2f1591f4feef89885c3f38c83d3a3474a5622cf8cd3de1b93849603",
        "Height": 100,
        "Transactions": [
            "37e017cb9de93aa93ef817e82c555812a0a6d5c3f7d6c521c7808a5a77fc93c7"
        ]
    }
}
```

#### 21. getnetworkid

获取 network id.

#### Example

Request:

```
{
  "jsonrpc": "2.0",
  "method": "getnetworkid",
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
  "result": 1
}
```

#### 22. getgrantong

获取 grant ong.

#### Example

Request:

```
{
  "jsonrpc": "2.0",
  "method": "getgrantong",
  "params": ["AKDFapcoUhewN9Kaj6XhHusurfHzUiZqUA"],
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
  "result": 4995625
}
```

## 错误代码

错误码定义

| 字段 | 类型 | 定义 |
| :--- | :--- | :--- |
| 0 | int64 | 成功 |
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
