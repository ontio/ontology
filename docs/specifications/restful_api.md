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
| get_unclaimong | GET /api/v1/unclaimong/:addr |
| get_mempooltxstate | GET /api/v1/mempool/txstate/:hash |
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

### 18 get_unclaimong

Get unclaimong

GET
```
/api/v1/unclaimong
```
#### Request Example:
```
curl -i http://localhost:20334/api/v1/unclaimong/:addr
```
#### Response
```
{
    "Action": "getunclaimong",
    "Desc": "SUCCESS",
    "Error": 0,
    "Result": "204957950400000",
    "Version": "1.0.0"
}
```

### 19 get_mempooltxstate

Query the transaction state in the memory pool.

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
