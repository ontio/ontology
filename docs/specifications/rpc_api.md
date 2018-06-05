# ONT Rpc Api

* [Introduction](#introduction)
* [Rpc Api List](#rpc-api-list)
* [Error Code](#error-code)

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
| result | object | program execution result |

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
| NextBookkeeper | Address | Accounting contract hash value for the next block |
| Bookkeepers | []*crypto.PubKey ||
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

## Rpc Api List

| Method | Parameters | Description | Note |
| :---| :---| :---| :---|
| getbestblockhash |  | get the hash of the highest height block in the main chain |  |
| getblock | height or blockhash,[verbose] | get block by block height or block hash | verbose can be 0 or 1,response is different |
| getblockcount |  | get the number of blocks |  |
| getblockhash | height | get block hash by block height |  |
| getconnectioncount|  | get the current number of connections for the node |  |
| getgenerateblocktime|  | The time required to create a new block |  |
| getrawtransaction | transactionhash | Returns the corresponding transaction information based on the specified hash value. |  |
| sendrawtransaction | hex,preExec | Broadcast transaction. | Serialized signed transactions constructed in the program into hexadecimal strings |
| getstorage | script_hash | Returns the stored value according to the contract script hashes and stored key. |  |
| getversion |  | Get the version information of the query node |  |
| getblocksysfee |  | According to the specified index, return the system fee before the block. |  |
| getcontractstate | script_hash,[verbose] | According to the contract script hash, query the contract information. |  |
| getmempooltxstate | tx_hash | Query the transaction status in the memory pool. |  |
| getsmartcodeevent |  | Get smartcode event |  |
| getblockheightbytxhash | tx_hash | get blockheight of txhash|  |
| getbalance | address | return balance of base58 account address. |  |
| getmerkleproof | tx_hash | return merkle_proof |  |
| getgasprice |  | return gasprice |  |
| getallowance | asset, from, to | return allowance |  |
| getunclaimong | address | return unclaimong |  |
| getblocktxsbyheight | height | return tx hashes |  |

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

The time required to create a new block

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
  "result": 6
}
```

#### 7. getrawtransaction

Returns the corresponding transaction information based on the specified hash value.

#### Parameter instruction

txid: transaction ID

Verbose: Optional parameter, the default value of verbose is 0, when verbose is 0, it returns the transaction serialized information, which is represented by a hexadecimal string. To get detailed information from it, you need to call the SDK to deserialize. When verbose is 1, the detailed information of the corresponding transaction is returned, which is represented by a JSON format string.

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

send transaction.

#### Parameter instruction

Hex: Serialized signed transactions constructed in the program into hexadecimal strings.Building the parameter,please refer to TestInvokefunction in ontology/http/func_test.go.
PreExec : set 1 if want prepare exec smartcontract

How to build the parameter?

Take the "AddAttribute" in the IdContract contract as an example

1. build parameter

```
acct := account.Open(account.WALLET_FILENAME, []byte("passwordtest"))
acc, err := acct.GetDefaultAccount()
pubkey := keypair.SerializePublicKey(acc.PubKey())
funcName := "AddAttribute"
paras := []interface{}{[]byte("did:ont:" + acc.Address.ToBase58()),[]byte("key1"),[]byte("bytes"),[]byte("value1"),pubkey}
builder := neovm.NewParamsBuilder(new(bytes.Buffer))
err = BuildSmartContractParamInter(builder, []interface{}{funcName, params})
codeParams := builder.ToArray()
op_verify,_ := common.HexToBytes("69")
codeaddress,_ := common.HexToBytes("8055b362904715fd84536e754868f4c8d27ca3f6")
codeParams = BytesCombine(codeParams,op_verify)
codeParams = BytesCombine(codeParams,codeaddress)

func BytesCombine(pBytes ...[]byte) []byte {
	len := len(pBytes)
	s := make([][]byte, len)
	for index := 0; index < len; index++ {
		s[index] = pBytes[index]
	}
	sep := []byte("")
	return bytes.Join(s, sep)
}
```
funcName:the smartcontract function name to be called, params: contract function required parameters, codeAddress: smartcontract address

2. build transaction
```
tx := utils.NewInvokeTransaction(vmtypes.VmCode{
		VmType: vmtypes.NEOVM,
		Code:   codeParams,
	})
	tx.Nonce = uint32(time.Now().Unix())
```

3. sign transaction

```
hash := tx.Hash()
sign, _ := signature.Sign(acc.PrivateKey, hash[:])
tx.Sigs = append(tx.Sigs, &ctypes.Sig{
    PubKeys: []keypair.PublicKey{acc.PublicKey},
    M:       1,
    SigData: [][]byte{sign},
})
```

4. Convert transactions to hexadecimal strings
```
txbf := new(bytes.Buffer)
err = tx.Serialize(txbf);
common.ToHexString(txbf.Bytes())
```

Related struct
```
type Transaction struct {
	Version    byte
	TxType     TransactionType
	Nonce      uint32
	Payload    Payload
	Attributes []*TxAttribute
	Fee        []*Fee
	NetWorkFee common.Fixed64
	Sigs       []*Sig

	hash *common.Uint256
}

type Sig struct {
	PubKeys []keypair.PublicKey
	M       uint8
	SigData [][]byte
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
    "jsonpc": "2.0",
    "result": "498db60e96828581eff991c58fa46abbfd97d2f4a4f9915a11f85c54f2a2fedf"
}
```

> Note:result is txhash

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

or

```
{
    "desc": "SUCCESS",
    "error": 0,
    "id": 1,
    "jsonpc": "2.0",
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

> Note: If params is a number, the response result will be the txhash list. If params is txhash, the response result will be smartcode event.

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
verbose: Optional parameter, the default value of verbose is 0, when verbose is 0, it returns the contract serialized information, which is represented by a hexadecimal string. To get detailed information from it, you need to call the SDK to deserialize. When verbose is 1, the detailed information of the corresponding contract is returned, which is represented by a JSON format string.

#### Example

Request:

```
{
  "jsonrpc": "2.0",
  "method": "getcontractstate",
  "params": ["fff49c809d302a2956e9dc0012619a452d4b846c",1],
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

Query the transaction state in the memory pool.

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
              	"State": [{
              		"Type": 1,
              		"Height": 342,
              		"ErrCode": 0
              	}, {
              		"Type": 0,
              		"Height": 0,
              		"ErrCode": 0
              	}],
              	"Transaction": {
              		"Nonce": 3383369367,
              		"Payer": "0100aeb10ff2919bfe14dc432899d3b649893119",
              		"TxType": 208,
              		"Version": 0,
              		"Sigs": [{
              			"PubKeys": ["120202d3d048aca7bdee582a611d0b8acc45642950dc6167aee63abbdcd1a5781c6319"],
              			"SigData": ["01af5becdd6ef2314985fdf510f63e080ee24c84f7e45da813ada165a04054666fa648b0458e3bb81512d30c6dd8efc2e5ddb3561a79fc5cccdb620e550fda563d"],
              			"M": 1
              		}],
              		"Attributes": [{
              			"Usage": 0,
              			"Data": "33623432313634662d336233392d346164302d623633382d363063363763343333613161"
              		}],
              		"GasPrice": 0,
              		"Payload": {
              			"NeedStorage": true,
              			"Email": "email",
              			"Description": "desp",
              			"VmType": 128,
              			"CodeVersion": "v1.0",
              			"Author": "author",
              			"Code": "0115c56b6c766b00527ac46c766b51527ac4616c766b00c304696e6974876c766b54527ac46c766b54c36411006165af026c766b55527ac46278026c766b00c30a6d696e74546f6b656e73876c766b56527ac46c766b56c36411006165e6046c766b55527ac4624a026c766b00c30b746f74616c537570706c79876c766b57527ac46c766b57c3641100616513096c766b55527ac4621b026c766b00c3046e616d65876c766b58527ac46c766b58c3641100616506026c766b55527ac462f3016c766b00c30673796d626f6c876c766b59527ac46c766b59c36411006165ed016c766b55527ac462c9016c766b00c306656e6449636f876c766b5a527ac46c766b5ac36411006165f7066c766b55527ac4629f016c766b00c3087472616e73666572876c766b5b527ac46c766b5bc3647100616c766b51c3c0539c009c6c766b5f527ac46c766b5fc3640e00006c766b55527ac4625c016c766b51c300c36c766b5c527ac46c766b51c351c36c766b5d527ac46c766b51c352c36c766b5e527ac46c766b5cc36c766b5dc36c766b5ec3615272655a086c766b55527ac46213016c766b00c30962616c616e63654f66876c766b60527ac46c766b60c3644d00616c766b51c3c0519c009c6c766b0112527ac46c766b0112c3640e00006c766b55527ac462cd006c766b51c300c36c766b0111527ac46c766b0111c36165360a6c766b55527ac462aa006c766b00c308646563696d616c73876c766b0113527ac46c766b0113c36411006165ab006c766b55527ac4627c006165870d6c766b52527ac46165fc0e6c766b53527ac46c766b53c300907c907ca1630e006c766b52c3c000a0620400006c766b0114527ac46c766b0114c364300061616c766b52c36c766b53c3617c06726566756e6453c168124e656f2e52756e74696d652e4e6f746966796161006c766b55527ac46203006c766b55c3616c756600c56b0943505820546f6b656e616c756600c56b03435058616c756600c56b58616c756653c56b616168164e656f2e53746f726167652e476574436f6e746578740b746f74616c537570706c79617c680f4e656f2e53746f726167652e4765746c766b00527ac46c766b00c3c000a06c766b51527ac46c766b51c3640e00006c766b52527ac462fa016168164e656f2e53746f726167652e476574436f6e746578746114b30f23279d728063e8b7414bc641a20ef56c6d05070000e6f4573466615272680f4e656f2e53746f726167652e5075746161006114b30f23279d728063e8b7414bc641a20ef56c6d05070000e6f4573466615272087472616e7366657254c168124e656f2e52756e74696d652e4e6f74696679616168164e656f2e53746f726167652e476574436f6e746578746114b30f23279d728063e8b7414bc641a20ef56c6d050700c06cc92ade0f615272680f4e656f2e53746f726167652e5075746161006114b30f23279d728063e8b7414bc641a20ef56c6d050700c06cc92ade0f615272087472616e7366657254c168124e656f2e52756e74696d652e4e6f74696679616168164e656f2e53746f726167652e476574436f6e7465787461140100aeb10ff2919bfe14dc432899d3b6498931190700a2333d4f3b6c615272680f4e656f2e53746f726167652e50757461610061140100aeb10ff2919bfe14dc432899d3b6498931190700a2333d4f3b6c615272087472616e7366657254c168124e656f2e52756e74696d652e4e6f74696679616168164e656f2e53746f726167652e476574436f6e746578740b746f74616c537570706c7904a2aef725615272680f4e656f2e53746f726167652e50757461516c766b52527ac46203006c766b52c3616c75665ec56b616165780a6c766b00527ac46c766b00c3c0009c6c766b59527ac46c766b59c3640f0061006c766b5a527ac46284026168184e656f2e426c6f636b636861696e2e4765744865696768746168184e656f2e426c6f636b636861696e2e4765744865616465726168174e656f2e4865616465722e47657454696d657374616d706c766b51527ac46c766b51c304d0396e5a946c766b52527ac46165610b6c766b53527ac46c766b00c36c766b53c36c766b52c361527265710e6c766b54527ac46c766b54c3009c6c766b5b527ac46c766b5bc3640f0061006c766b5a527ac462d2016c766b52c36165ab066c766b55527ac46c766b55c3009c6c766b5c527ac46c766b5cc3643a0061616c766b00c36c766b54c3617c06726566756e6453c168124e656f2e52756e74696d652e4e6f7469667961006c766b5a527ac46275016c766b00c36c766b54c36c766b55c3615272659a066c766b56527ac46c766b56c3009c6c766b5d527ac46c766b5dc3640f0061006c766b5a527ac46237016168164e656f2e53746f726167652e476574436f6e746578746c766b00c3617c680f4e656f2e53746f726167652e4765746c766b57527ac46168164e656f2e53746f726167652e476574436f6e746578746c766b00c36c766b56c36c766b57c393615272680f4e656f2e53746f726167652e507574616168164e656f2e53746f726167652e476574436f6e746578740b746f74616c537570706c79617c680f4e656f2e53746f726167652e4765746c766b58527ac46168164e656f2e53746f726167652e476574436f6e746578740b746f74616c537570706c796c766b56c36c766b58c393615272680f4e656f2e53746f726167652e5075746161006c766b00c36c766b56c3615272087472616e7366657254c168124e656f2e52756e74696d652e4e6f7469667961516c766b5a527ac46203006c766b5ac3616c756655c56b616114f9a28e064678c6ad2d90f7b1d6886d3a6fa34f896168184e656f2e52756e74696d652e436865636b5769746e657373009c6c766b52527ac46c766b52c3640e00006c766b53527ac46249016168164e656f2e53746f726167652e476574436f6e746578740b746f74616c537570706c79617c680f4e656f2e53746f726167652e4765746c766b00527ac404a2085a286c766b00c3946c766b51527ac46c766b51c300a16c766b54527ac46c766b54c3640f0061006c766b53527ac462d6006168164e656f2e53746f726167652e476574436f6e746578740b746f74616c537570706c7904a2085a28615272680f4e656f2e53746f726167652e507574616168164e656f2e53746f726167652e476574436f6e746578746114f9a28e064678c6ad2d90f7b1d6886d3a6fa34f896c766b51c3615272680f4e656f2e53746f726167652e5075746161006114f9a28e064678c6ad2d90f7b1d6886d3a6fa34f896c766b51c3615272087472616e7366657254c168124e656f2e52756e74696d652e4e6f7469667961516c766b53527ac46203006c766b53c3616c756651c56b616168164e656f2e53746f726167652e476574436f6e746578740b746f74616c537570706c79617c680f4e656f2e53746f726167652e4765746c766b00527ac46203006c766b00c3616c75665bc56b6c766b00527ac46c766b51527ac46c766b52527ac4616c766b52c300a16c766b55527ac46c766b55c3640e00006c766b56527ac46205026c766b00c36168184e656f2e52756e74696d652e436865636b5769746e657373009c6c766b57527ac46c766b57c3640e00006c766b56527ac462c9016c766b00c36c766b51c39c6c766b58527ac46c766b58c3640e00516c766b56527ac462a4016168164e656f2e53746f726167652e476574436f6e746578746c766b00c3617c680f4e656f2e53746f726167652e4765746c766b53527ac46c766b53c36c766b52c39f6c766b59527ac46c766b59c3640e00006c766b56527ac46247016c766b53c36c766b52c39c6c766b5a527ac46c766b5ac3643b006168164e656f2e53746f726167652e476574436f6e746578746c766b00c3617c68124e656f2e53746f726167652e44656c657465616241006168164e656f2e53746f726167652e476574436f6e746578746c766b00c36c766b53c36c766b52c394615272680f4e656f2e53746f726167652e507574616168164e656f2e53746f726167652e476574436f6e746578746c766b51c3617c680f4e656f2e53746f726167652e4765746c766b54527ac46168164e656f2e53746f726167652e476574436f6e746578746c766b51c36c766b54c36c766b52c393615272680f4e656f2e53746f726167652e50757461616c766b00c36c766b51c36c766b52c3615272087472616e7366657254c168124e656f2e52756e74696d652e4e6f7469667961516c766b56527ac46203006c766b56c3616c756652c56b6c766b00527ac4616168164e656f2e53746f726167652e476574436f6e746578746c766b00c3617c680f4e656f2e53746f726167652e4765746c766b51527ac46203006c766b51c3616c756652c56b616168164e656f2e53746f726167652e476574436f6e746578740669636f4e656f617c680f4e656f2e53746f726167652e4765746c766b00527ac46c766b00c36c766b51527ac46203006c766b51c3616c756653c56b6c766b00527ac4616c766b00c3009f6310006c766b00c303008d27a0620400516c766b51527ac46c766b51c3640f0061006c766b52527ac4621400610500e87648176c766b52527ac46203006c766b52c3616c75665ec56b6c766b00527ac46c766b51527ac46c766b52527ac4616c766b51c30400e1f505966c766b52c3956c766b53527ac46168164e656f2e53746f726167652e476574436f6e746578740b746f74616c537570706c79617c680f4e656f2e53746f726167652e4765746c766b54527ac46168164e656f2e53746f726167652e476574436f6e746578740669636f4e656f617c680f4e656f2e53746f726167652e4765746c766b55527ac404a2085a286c766b54c3946c766b56527ac46c766b56c300a16c766b57527ac46c766b57c3643a0061616c766b00c36c766b51c3617c06726566756e6453c168124e656f2e52756e74696d652e4e6f7469667961006c766b58527ac46281016c766b56c36c766b53c39f6c766b59527ac46c766b59c3640c01616c766b56c36c766b5a527ac46c766b5ac36c766b52c3966c766b5b527ac46c766b51c36c766b5bc30400e1f50595946c766b5c527ac4616c766b00c36c766b5cc3617c06726566756e6453c168124e656f2e52756e74696d652e4e6f74696679616168164e656f2e53746f726167652e476574436f6e746578740b746f74616c537570706c79617c680f4e656f2e53746f726167652e4765746c766b5d527ac46168164e656f2e53746f726167652e476574436f6e746578740b746f74616c537570706c796c766b5dc36c766b5ac393615272680f4e656f2e53746f726167652e507574616c766b5bc30400e1f505956a51527ac46c766b5bc36c766b52c3956c766b53527ac4616c766b55c36c766b51c3936c766b55527ac46168164e656f2e53746f726167652e476574436f6e746578740669636f4e656f6c766b55c3615272680f4e656f2e53746f726167652e507574616c766b53c36c766b58527ac46203006c766b58c3616c756657c56b6161682953797374656d2e457865637574696f6e456e67696e652e476574536372697074436f6e7461696e65726c766b00527ac46c766b00c361681d4e656f2e5472616e73616374696f6e2e4765745265666572656e6365736c766b51527ac4616c766b51c36c766b52527ac4006c766b53527ac4629e006c766b52c36c766b53c3c36c766b54527ac4616c766b54c36168154e656f2e4f75747075742e4765744173736574496461209b7cffdaa674beae0f930ebe6085af9093e5fe56b34a5c220ccdcf6efc336fc59c6c766b55527ac46c766b55c3642d006c766b54c36168184e656f2e4f75747075742e476574536372697074486173686c766b56527ac4622c00616c766b53c351936c766b53527ac46c766b53c36c766b52c3c09f6359ff006c766b56527ac46203006c766b56c3616c756651c56b6161682d53797374656d2e457865637574696f6e456e67696e652e476574457865637574696e67536372697074486173686c766b00527ac46203006c766b00c3616c756658c56b6161682953797374656d2e457865637574696f6e456e67696e652e476574536372697074436f6e7461696e65726c766b00527ac46c766b00c361681a4e656f2e5472616e73616374696f6e2e4765744f7574707574736c766b51527ac4006c766b52527ac4616c766b51c36c766b53527ac4006c766b54527ac462ce006c766b53c36c766b54c3c36c766b55527ac4616c766b55c36168184e656f2e4f75747075742e47657453637269707448617368616505ff907c907c9e6346006c766b55c36168154e656f2e4f75747075742e4765744173736574496461209b7cffdaa674beae0f930ebe6085af9093e5fe56b34a5c220ccdcf6efc336fc59c620400006c766b56527ac46c766b56c3642d00616c766b52c36c766b55c36168134e656f2e4f75747075742e47657456616c7565936c766b52527ac461616c766b54c351936c766b54527ac46c766b54c36c766b53c3c09f6329ff6c766b52c36c766b57527ac46203006c766b57c3616c756659c56b6c766b00527ac46c766b51527ac46c766b52527ac46c766b53527ac4616168164e656f2e53746f726167652e476574436f6e746578746c766b00c36c766b53c37e617c680f4e656f2e53746f726167652e4765746c766b54527ac46c766b52c36c766b54c3946c766b55527ac46c766b55c300a0009c6c766b56527ac46c766b56c3643a0061616c766b00c36c766b51c3617c06726566756e6453c168124e656f2e52756e74696d652e4e6f7469667961006c766b57527ac462f5006c766b55c36c766b51c39f6c766b58527ac46c766b58c364880061616c766b00c36c766b51c36c766b55c394617c06726566756e6453c168124e656f2e52756e74696d652e4e6f74696679616168164e656f2e53746f726167652e476574436f6e746578746c766b00c36c766b53c37e6c766b55c36c766b54c393615272680f4e656f2e53746f726167652e507574616c766b55c36c766b57527ac46256006168164e656f2e53746f726167652e476574436f6e746578746c766b00c36c766b53c37e6c766b51c36c766b54c393615272680f4e656f2e53746f726167652e507574616c766b51c36c766b57527ac46203006c766b57c3616c756658c56b6c766b00527ac46c766b51527ac46c766b52527ac4616c766b52c3009f6c766b53527ac46c766b53c3643a0061616c766b00c36c766b51c3617c06726566756e6453c168124e656f2e52756e74696d652e4e6f7469667961006c766b54527ac46231016c766b52c302100ea0009c6c766b55527ac46c766b55c3643b00616c766b00c36c766b51c3040084d717056361705f31615379517955727551727552795279547275527275659bfd6c766b54527ac462df006c766b52c302201ca0009c6c766b56527ac46c766b56c3643b00616c766b00c36c766b51c30400ca9a3b056361705f326153795179557275517275527952795472755272756549fd6c766b54527ac4628d006c766b52c302302aa0009c6c766b57527ac46c766b57c3643c00616c766b00c36c766b51c305005ed0b200056361705f3361537951795572755172755279527954727552727565f6fc6c766b54527ac4623a0061616c766b00c36c766b51c3617c06726566756e6453c168124e656f2e52756e74696d652e4e6f7469667961006c766b54527ac46203006c766b54c3616c7566",
              			"Name": "name"
              		},
              		"Hash": "5f4ef20deeb22b58ba4085e6922c1f6fdf44556fb4d7aaac26d50191d0e50b7e",
              		"GasLimit": 10000000
              	}
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
        "ont": "2500",
        "ong": "0",
        "ong_appove": "0"
       }
}
```

#### 17. getmerkleproof

return merkle proof

#### Parameter instruction

hash: transaction hash

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
   "jsonpc":"2.0",
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

#### 18. getgasprice

return gasprice.


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
   "jsonpc":"2.0",
   "result":{
        "gasprice": 0,
        "height": 1
       }
}
```

#### 19. getallowance

return allowance.


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
   "jsonpc":"2.0",
   "result": "10"
}
```

#### 20. getunclaimong

return unclaimong.


#### Example

Request:

```
{
  "jsonrpc": "2.0",
  "method": "getunclaimong",
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
   "jsonpc":"2.0",
   "result": "204957950400000"
}
```

#### 21 getblocktxsbyheight

Get transactions by block height
return all transaction hash contained in the block corresponding to this height

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
   "jsonpc":"2.0",
   "result": {
        "Hash": "ea5e5219d2f1591f4feef89885c3f38c83d3a3474a5622cf8cd3de1b93849603",
        "Height": 100,
        "Transactions": [
            "37e017cb9de93aa93ef817e82c555812a0a6d5c3f7d6c521c7808a5a77fc93c7"
        ]
    }
}
```

## Error Code

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
