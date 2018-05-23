# Ontology Restful API

* [Introduction](#introduction)
* [Restful Api List](#restful-api-list)
* [Error Code](#error-code)

Restful Api List

| Method | url |
| :---| :---|
| get_gen_blk_time | GET /api/v1/node/generateblocktime |
| get_conn_count | GET /api/v1/node/connectioncount |
| get_blk_txs_by_height | GET /api/v1/block/transactions/height/:height |
| get_blk_by_height | GET /api/v1/block/details/height/:height?raw=0 |
| get_blk_by_hash | GET /api/v1/block/details/hash/:hash?raw=1 |
| get_blk_height | GET /api/v1/block/height |
| get_blk_hash | GET /api/v1/block/hash/:height |
| get_tx | GET /api/v1/transaction/:hash |
| get_storage | GET /api/v1/storage/:hash/:key|
| get_balance | GET /api/v1/balance/:addr |
| get_contract_state | GET /api/v1/contract/:hash |
| get_smtcode_evt_txs | GET /api/v1/smartcode/event/transactions/:height |
| get_smtcode_evts | GET /api/v1/smartcode/event/txhash/:hash |
| get_blk_hgt_by_txhash | GET /api/v1/block/height/txhash/:hash |
| get_merkle_proof | GET /api/v1/merkleproof/:hash|
| get_gasprice | GET /api/v1/gasprice|
| get_allowance | GET /api/v1/allowance/:asset/:from/:to |
| post_raw_tx | post /api/v1/transaction?preExec=0 |


## Introduction

This document describes the restful api format for the http/https used in the Onchain Ontology.

## Restful Api List

### Response parameters descri

| Field | Type | Description |
| :--- | :--- | :--- |
| Action | string | action name |
| Desc | string | description |
| Error | int64 | error code |
| Result | object | execute result |
| Version | string | version information |

### 1. get_gen_blk_time

Get the generate block time

##### GET

```
/api/v1/node/generateblocktime
```
#### Request Example:

```
curl -i http://server:port/api/v1/node/generateblocktime
```

#### Response example:

```
{
    "Action": "getgenerateblocktime",
    "Desc": "SUCCESS"
    "Error": 0,
    "Result": 6,
    "Version": "1.0.0"
}
```
### 2 get_conn_count

Get the number of connected node


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
### 3 get_blk_txs_by_height

Get transactions by block height
return all transaction hash contained in the block corresponding to this height

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
### 4 get_blk_by_height

Get the block by block height
return block details based on block height
if raw=1 return serialized block
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
### 5 get_blk_by_hash

Get block by blockhash
return block details based on block hash,if raw=1 return serialized block

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

### 6 get_blk_height

Get the current block height


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

### 7 get_blk_hash

Get blockhash by block height

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

### 8 get_tx

get transaction by transaction hash,if raw=1 return serialized transaction

GET

```
/api/v1/transaction/:hash?raw=0
```

####Request Example:

```
curl -i http://server:port/api/v1/transaction/c5e0d387c6a97aef12f1750840d24b53d9fe7f22f16c7b7703d4a93a28370baa
```
#### Response Example:

```
{
    "Action": "gettransaction",
    "Desc": "SUCCESS",
    "Error": 0,
    "Result": {
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
    },
    "Version": "1.0.0"
}
```

### 9 post_raw_tx

send transaction. set preExec=1 if want prepare exec smartcontract

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
> Result: txhash

### 10 get_storage

Returns the stored value according to the contract script hashes and stored key.

GET
```
/api/v1/storage/:hash/:key
```
Request Example
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
> Result:Returns the stored value according to the contract script hashes and stored key.

### 11 get_balance

return balance of base58 account address.

GET
```
/api/v1/balance/:addr
```
> addr: Base58 encoded account address

Request Example
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
        "ong": "0",
        "ong_appove": "0"
    },
    "Version": "1.0.0"
}
```
### 12 get_contract_state

According to the contract script hash, query the contract information.

GET

```
/api/v1/contract/:hash
```

#### Request Example:

```
curl -i http://server:port/api/v1/contract/fff49c809d302a2956e9dc0012619a452d4b846c
```

#### Response Example:

```
{
    "Action": "getcontract",
    "Desc": "SUCCESS",
    "Error": 0,
    "Version": "1.0.0",
    "Result": {
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

#### 13 get_smtcode_evt_txs

get smart contract event txhash list by height
Get a list of transaction hash with smartevent based on height

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
        "592d83c739d9d167b74b385161fee09bfe820eae5bc4a69411f8e00f4847b833"
    ],
    "Version": "1.0.0"
}
```
> Note: result is the txHash list.

### 14 get_smtcode_evts

get contract event by txhash

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
### 15 get_blk_hgt_by_txhash

Get block height by transaction hash

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

### 16 get_merkle_proof

get merkle proof

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

### 17 get_gasprice

Get gasprice

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

### 18 get_allowance

Get allowance

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

## Error Code

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
