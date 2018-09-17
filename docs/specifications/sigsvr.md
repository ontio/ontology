# Ontology Signature Server Tutorials

[English|[中文](sigsvr_CN.md)]

Ontology Signature Server - sigsvr is a rpc server for signing transactions.

* [Ontology Signature Server Tutorials](#ontology-signature-server-tutorials)
	* [1. Signature Service Startup](#1-signature-service-startup)
		* [1.1 The Parameters of Signature Service Startup](#11-the-parameters-of-signature-service-startup)
		* [1.2 Import wallet account](#12-import-wallet-account)
			* [1.2.1 Import wallet account parameters](#121-import-wallet-account-parameters)
		* [1.3 Startup](#13-startup)
	* [2. Signature Service Method](#2-signature-service-method)
		* [2.1  Signature Service Calling Method](#21-signature-service-calling-method)
		* [2.2 Signature for Data](#22-signature-for-data)
		* [2.3 Signature for Raw Transactions](#23-signature-for-raw-transactions)
		* [2.4 Multiple Signature for Raw Transactions](#24-multiple-signature-for-raw-transactions)
		* [2.5 Signature of Transfer Transaction](#25-signature-of-transfer-transaction)
		* [2.6 Native Contract Invokes Signature](#26-native-contract-invokes-signature)
			* [Example1:  Constructing  transfer transaction](#example1-constructing-transfer-transaction)
			* [Example2: Constructing  withdraw ONG transaction](#example2-constructing-withdraw-ong-transaction)
		* [2.7 NeoVM Contract Invokes Signature](#27-neovm-contract-invokes-signature)
		* [2.8 NeoVM Contract Invokes By ABI Signature](#28-neovm-contract-invokes-by-abi-signature)
		* [2.9 Create Account](#29-create-account)
		* [2.10 ExportAccount](#210-exportaccount)

## 1. Signature Service Startup

### 1.1 The Parameters of Signature Service Startup

--loglevel
The loglevel parameter is used to set the log level for the sigsvr output. Sigsvr supports 7 different log levels - 0:Trace 1:Debug 2:Info 3:Warn 4:Error 5:Fatal 6:MaxLevel. The log level is from low to high, and the output log volume is from high to low. The default value is 2, which means that only output logs at the info level or higher level.

--walletdir
walletdir parameter specifies the directory for wallet data. The default value is "./wallet_data".

--cliaddress
The cliaddress parameter specifies which address is bound。The default value is 127.0.0.1, means only local machine's request can be accessed。If sigsvr need be accessed by other machine, please use local network address or 0.0.0.0。

--cliport
The port number to which the signature server is bound. The default value is 20000.

--abi
abi parameter specifies the abi file path when sigsvr starts. The default value is "./abi".

### 1.2 Import wallet account

Before startup sigsvr, should import wallet account.

#### 1.2.1 Import wallet account parameters

--walletdir
walletdir parameter specifies the directory for wallet data. The default value is "./wallet_data".

--wallet
wallet parameter specifies the path of wallet to import. The default value is "./wallet.dat".

**Import wallet account**

```
./sigsvr import
```

### 1.3 Startup

```
./sigsvr
```

## 2. Signature Service Method

The signature service currently supports signature for data, single signature and multi-signatures for raw transactions, constructing ONT/ONG transfer transactions and signing, constructing transactions that Native contracts can invoke and signing, and constructing transactions that NeoVM contracts can invoke and signing, and so on.

### 2.1  Signature Service Calling Method

The signature service is a json rpc server and adopts the POST method. The requested service path is:

```
http://localhost:20000/cli
```
Request structure:

```
{
    "qid":"XXX",    //Request ID， the response will bring the same qid
    "method":"XXX", //Requested method name
    "account":"XXX",//account for sign
    "pwd":"XXX",    //unlock password
    "params":{
    	//The request parameters that are filled in according to the request method
    }
}
```
Response structure:

```
{
    "qid": "XXX",   //Request ID
    "method": "XXX",//Requested method name
    "result": {     //Response result
        "signed_tx": "XXX"  //Signed transaction
    },
    "error_code": 0,//Error code，zero represents success, non-zero represents failure
    "error_info": ""//Error description
}
```

Error code:

Error code  | Error description
--------|-----------
1001 | Invalid http method
1002 | Invalid http request
1003 | Invalid request parameter
1004 | Unsupported method
1005 | Account is locked
1006 | Invalid transactions
1007 | ABI is not found
1008 | ABI is not matched
9999 | Unknown error

### 2.2 Signature for Data

SigSvr can signature for any data. Note that data must be encode by hex string.

Method Name: sigdata

Request parameters:

```
{
    "raw_data":"XXX"      //Unsigned data, Note that data must be encode by hex string.
}
```
Response result:

```
{
    "signed_data":"XXX"   //Signed data, Note that data was encoded by hex string.
}
```
Examples:

Request:

```
{
    "qid":"t",
    "method":"sigdata",
    "account":"XXX",
    "pwd":"XXX",
    "params":{
    	"raw_data":"48656C6C6F20776F726C64" //Hello world
    }
}
```
Response:

```
{
    "qid": "t",
    "method": "sigdata",
    "result": {
        "signed_data": "cab96ef92419df915902817b2c9ed3f6c1c4956b3115737f7c787b03eed3f49e56547f3117867db64217b84cd6c6541d7b248f23ceeef3266a9a0bd6497260cb"
    },
    "error_code": 0,
    "error_info": ""
}
```

### 2.3 Signature for Raw Transactions

Method Name: sigrawtx

Request parameters:

```
{
    "raw_tx":"XXX"      //Unsigned transaction
}
```
Response result:

```
{
    "signed_tx":"XXX"   //Signed transaction
}
```
Examples:

Request:
```
{
    "qid":"1",
    "method":"sigrawtx",
    "account":"XXX",
    "pwd":"XXX",
    "params":{
    	"raw_tx":"00d14150175b000000000000000000000000000000000000000000000000000000000000000000000000ff4a0000ff00000000000000000000000000000000000001087472616e736665722a0101d4054faaf30a43841335a2fbc4e8400f1c44540163d551fe47ba12ec6524b67734796daaf87f7d0a0000"
    }
}
```
Response:

```
{
    "qid": "1",
    "method": "sigrawtx",
    "result": {
        "signed_tx": "00d14150175b00000000000000000000000000000000011e68f7bf0aaba1f18213639591f932556eb674ff4a0000ff00000000000000000000000000000000000001087472616e736665722a0101d4054faaf30a43841335a2fbc4e8400f1c44540163d551fe47ba12ec6524b67734796daaf87f7d0a000101231202026940ba3dba0a385c44e4a187af75a34e281b96200430db2cbc688a907e5fb54501014101d3998581639ff873f4e63936ae63d6ccd56d0f756545ba985951073f293b07507e2f1a342654c8ad28d092dd6d8250a0b29f5c00c7866df3c4df1cff8f00c6bb"
    },
    "error_code": 0,
    "error_info": ""
}
```

### 2.4 Multiple Signature for Raw Transactions

Since the private key is in the hands of different people, the multi-signature method needs to be called multiple times.

Method Name: sigmutilrawtx

Request parameters:
```
{
    "raw_tx":"XXX", //Unsigned transaction
    "m":xxx         //The minimum number of signatures required for multiple signatures
    "pub_keys":[
        //Public key list of signature
    ]
}
```
Response result:

```
{
    "signed_tx":"XXX" //Signed transaction
}
```
Examples:

Request:

```
{
    "qid":"1",
    "method":"sigmutilrawtx",
    "account":"XXX",
    "pwd":"XXX",
    "params":{
    	"raw_tx":"00d12454175b000000000000000000000000000000000000000000000000000000000000000000000000ff4a0000ff00000000000000000000000000000000000001087472616e736665722a01024ce71f6cc6c0819191e9ec9419928b183d6570012fb5cfb78c651669fac98d8f62b5143ab091e70a0000",
    	"m":2,
    	"pub_keys":[
    	    "1202039b196d5ed74a4d771ade78752734957346597b31384c3047c1946ce96211c2a7",
    	    "120203428daa06375b8dd40a5fc249f1d8032e578b5ebb5c62368fc6c5206d8798a966"
    	]
    }
}
```

Response:

```
{
    "qid": "1",
    "method": "sigmutilrawtx",
    "result": {
        "signed_tx": "00d12454175b00000000000000000000000000000000024ce71f6cc6c0819191e9ec9419928b183d6570ff4a0000ff00000000000000000000000000000000000001087472616e736665722a01024ce71f6cc6c0819191e9ec9419928b183d6570012fb5cfb78c651669fac98d8f62b5143ab091e70a000102231202039b196d5ed74a4d771ade78752734957346597b31384c3047c1946ce96211c2a723120203428daa06375b8dd40a5fc249f1d8032e578b5ebb5c62368fc6c5206d8798a9660201410166ec86a849e011e4c18d11d64ca0afbeaebf8a4c975be1eab4dcb8795abef5c908647294822ddaadaf2e4ae2432c9c8d143ceba8fa6355d6dfe59f846ac5a41a"
    },
    "error_code": 0,
    "error_info": ""
}
```
### 2.5 Signature of Transfer Transaction

In order to simplify the signature process of transfer transaction, a transfer transaction structure function is provided. When transferring, only the transfer parameters need to be provided.

Method Name: sigtransfertx
Request parameters:
```
{
    "gas_price":XXX,  //gasprice
    "gas_limit":XXX,  //gaslimit
    "asset":"ont",    //asset: ont or ong
    "from":"XXX",     //Payment account
    "to":"XXX",       //Receipt address
    "amount":"XXX"    //transfer amount. Note that since the precision of ong is 9, it is necessary to multiply the actual transfer amount by 1000000000 when making ong transfer.
}
```
Response result:

```
{
    "signed_tx":XXX     //Signed transaction
}
```

Examples:

Request:
```
{
    "qid":"t",
    "method":"sigtransfertx",
    "account":"XXX",
    "pwd":"XXX",
    "params":{
    	"gas_price":0,
    	"gas_limit":20000,
    	"asset":"ont",
    	"from":"ATACcJPZ8eECdWS4ashaMdqzhywpRTq3oN",
    	"to":"AeoBhZtS8AmGp3Zt4LxvCqhdU4eSGiK44M",
    	"amount":"10"
    }
}
```

Response:

```
{
    "qid": "t",
    "method": "sigtransfertx",
    "result": {
        "signed_tx": "00d1184a175b000000000000000050c3000000000000011e68f7bf0aaba1f18213639591f932556eb674ff4a0000ff00000000000000000000000000000000000001087472616e736665722a0102ae97551187192cdae14052c503b5e64b32013d01397cafa8cc71ae9e555e439fc0f0a5ded12a2a0a000101231202026940ba3dba0a385c44e4a187af75a34e281b96200430db2cbc688a907e5fb545010141016a0097dbe272fd61384d95aaeb02e9460e18078c9f4cd524d67ba431033b10d20de72d1bc819bf6e71c6c17116f4a7a9dfc9b395738893425d684faa30efea9c"
    },
    "error_code": 0,
    "error_info": ""
}
```

sigtransfertx method use the signer account to pay network fee by default, if you want to use other account to payer the fee, please use payer parameter to specifies。
Note that if specifies payer parameter, don't forget to use sigrawtx method to sign the transaction output by sigtransfertx with the fee payer's account.

Examples:
```
{
    "gas_price":XXX,  //gasprice
    "gas_limit":XXX,  //gaslimit
    "asset":"ont",    //asset: ont or ong
    "from":"XXX",     //Payment account
    "to":"XXX",       //Receipt address
    "payer":"XXX",    //The fee payer's account address
    "amount":XXX      //transfer amount. Note that since the precision of ong is 9, it is necessary to multiply the actual transfer amount by 1000000000 when making ong transfer.
}
```

### 2.6 Native Contract Invokes Signature

The Native contract invocation transaction is constructed and signed according to the ABI.


Note:
When sigsvr starts, it will default seek the native contract abi under "./abi" in the current directory. If there is no abi for this contract in the naitve directory, then it will return a 1007 error. Can use --abi parameter to change the abi seeking path.


Method Name: signativeinvoketx

Request parameters:

```
{
    "gas_price":XXX,    //gasprice
    "gas_limit":XXX,    //gaslimit
    "address":"XXX",    //The address that invokes native contract
    "method":"XXX",     //The method that invokes native contract
    "version":0,        //The version that invokes native contract
    "params":[
        //The parameters of the Native contract are constructed according to the ABI of calling method. All values ​​are string type.
    ]
}
```
Response result:
```
{
    "signed_tx":XXX     //Signed Transaction
}
```

#### Example1:  Constructing  transfer transaction

Request:

```
{
    "Qid":"t",
    "Method":"signativeinvoketx",
    "account":"XXX",
    "pwd":"XXX",
    "Params":{
    	"gas_price":0,
    	"gas_limit":20000,
    	"address":"0100000000000000000000000000000000000000",
    	"method":"transfer",
    	"version":0,
    	"params":[
    		[
    			[
    			"ATACcJPZ8eECdWS4ashaMdqzhywpRTq3oN",
    			"AeoBhZtS8AmGp3Zt4LxvCqhdU4eSGiK44M",
    			"1000"
    			]
    		]
    	]
    }
}
```
Response:

```
{
    "qid": "t",
    "method": "signativeinvoketx",
    "result": {
        "signed_tx": "00d161b7315b000000000000000050c3000000000000084c8f4060607444fc95033bd0a9046976d3a9f57300c66b147ce25d11becca9aa8f157e24e2a14fe100db73466a7cc814fc8a60f9a7ab04241a983817b04de95a8b2d4fb86a7cc802e8036a7cc86c51c1087472616e736665721400000000000000000000000000000000000000010068164f6e746f6c6f67792e4e61746976652e496e766f6b6500014140c4142d9e066fea8a68303acd7193cb315662131da3bab25bc1c6f8118746f955855896cfb433208148fddc0bed5a99dfde519fe063bbf1ff5e730f7ae6616ee02321035f363567ff82be6f70ece8e16378871128788d5a067267e1ec119eedc408ea58ac"
    },
    "error_code": 0,
    "error_info": ""
}
```

signativeinvoketx method use the signer account to pay network fee by default, if you want to use other account to payer the fee, please use payer parameter to specifies。
Note that if specifies payer parameter, don't forget to use sigrawtx method to sign the transaction output by signativeinvoketx with the fee payer's account.

Examples:

```
{
    "gas_price":XXX,    //gasprice
    "gas_limit":XXX,    //gaslimit
    "address":"XXX",    //The address that invokes NeoVM contract
    "payer":"XXX",      //The fee payer's account address
    "params":[
        //The parameters of the Native contract. All values are string type.
    ]
}
```
#### Example2: Constructing withdraw ONG transaction

``` json
{
	"Qid":"t",
	"Method":"signativeinvoketx",
	"account":"ARVVxBPGySL56CvSSWfjRVVyZYpNZ7zp48",	  //withdraw address
	"pwd":"XXX",
	"Params":{
		"gas_price":5000,
		"gas_limit":20000,
		"address":"0200000000000000000000000000000000000000",
		"method":"transferFrom",
		"version":0,
		"params":[
			"ARVVxBPGySL56CvSSWfjRVVyZYpNZ7zp48",	//withdraw address
			"AFmseVrdL9f9oyCzZefL9tG6UbvhUMqNMV",   //ONT contract address (in base58 style)
			"ARVVxBPGySL56CvSSWfjRVVyZYpNZ7zp48",  //ONG accept address. Note that accept address  can different with withdraw address
			"310860000000000"												//withdraw ong amount. Note that ONG has 9 decimals
		]
	}
}
```

### 2.7 NeoVM Contract Invokes Signature

The NeoVM parameter contract supports array, bytearray, string, int, and bool types. When constructing parameters, it is necessary to provide parameter types and parameter values. The parameter values use string types. Array is an array of objects and supports all types and quantities of NeoVM supported parameters.

Method Name: signeovminvoketx

Request parameters:
```
{
    "gas_price":XXX,    //gasprice
    "gas_limit":XXX,    //gaslimit
    "address":"XXX",    //The address that invokes NeoVM contract
    "params":[
        //The parameters of the Native contract. All values are string type.
    ]
}
```
Response result:
```
{
    "signed_tx":XXX     //Signed Transaction
}
```

Examples:
Request:

```
{
    "qid": "t",
    "method": "signeovminvoketx",
    "account":"XXX",
    "pwd":"XXX",
    "params": {
    	"gas_price": 0,
    	"gas_limit": 50000,
    	"address": "8074775331499ebc81ff785e299d406f55224a4c",
    	"version": 0,
    	"params": [
    		{
    			"type": "string",
    			"value": "Time"
    		},
    		{
    			"type": "array",
    			"value": [
    				{
    					"type": "string",
    					"value": ""
    				}
    			]
    		}
    	]
    }
}
```
Response:

```
{
    "qid": "t",
    "method": "signeovminvoketx",
    "result": {
        "signed_tx": "00d18f5e175b000000000000000050c3000000000000011e68f7bf0aaba1f18213639591f932556eb67480216700008074775331499ebc81ff785e299d406f55224a4c00080051c10454696d65000101231202026940ba3dba0a385c44e4a187af75a34e281b96200430db2cbc688a907e5fb54501014101b93bef619b4d7900b57f91e1810b268f9e10eb39fd563f23ce01323cde6273518000dc77d2d2231bc39428f1fa35d294990676015dbf6b4dfd2e6c9856034cc1"
    },
    "error_code": 0,
    "error_info": ""
}
```

signeovminvoketx method use the signer account to pay network fee by default, if you want to use other account to payer the fee, please use payer parameter to specifies。
Note that if specifies payer parameter, don't forget to use sigrawtx method to sign the transaction output by signeovminvoketx with the fee payer's account.

Examples:
```
{
    "gas_price":XXX,    //gasprice
    "gas_limit":XXX,    //gaslimit
    "address":"XXX",    //The address that invokes native contract
    "method":"XXX",     //The method that invokes native contract
    "version":0,        //The version that invokes native contract
    "payer":"XXX",      //The fee payer's account address
    "params":[
        //The parameters of the Native contract are constructed according to the ABI of calling method. All values are string type.
    ]
}
```

### 2.8 NeoVM Contract Invokes By ABI Signature

NeoVM contract invoke by abi transaction is constructed and signed according to the ABI, need the ABI of contract and invoke parameters.
Note that all value of parameters are string type.

Method Name: signeovminvokeabitx

Request parameters:

```
{
    "gas_price":XXX,    //gasprice
    "gas_limit":XXX,    //gaslimit
    "address":"XXX",    //The NeoVM contract address
    "params":[XXX],     //The parameters of the NeoVM contract are constructed according to the ABI of calling method. All values are string type.
    "contract_abi":XXX, //The ABI of contract
}
```
Response result:
```
{
    "signed_tx":XXX     //Signed Transaction
}
```

Examples:
Request:

```
{
    "qid": "t",
    "method": "signeovminvokeabitx",
    "account":"XXX",
    "pwd":"XXX",
    "params": {
    "gas_price": 0,
    "gas_limit": 50000,
    "address": "80b82b5e31ad8b7b750207ad80579b5296bf27e8",
    "method": "add",
    "params": ["10","10"],
    "contract_abi": {
        "hash": "0xe827bf96529b5780ad0702757b8bad315e2bb8ce",
        "entrypoint": "Main",
        "functions": [
            {
                "name": "Main",
                "parameters": [
                    {
                        "name": "operation",
                        "type": "String"
                    },
                    {
                        "name": "args",
                        "type": "Array"
                    }
                ],
                "returntype": "Any"
            },
            {
                "name": "Add",
                "parameters": [
                    {
                        "name": "a",
                        "type": "Integer"
                    },
                    {
                        "name": "b",
                        "type": "Integer"
                    }
                ],
                "returntype": "Integer"
            }
        ],
        "events": []
        }
    }
}
```

Response:
```
{
    "qid": "t",
    "method": "signeovminvokeabitx",
    "result": {
        "signed_tx": "00d16acd295b000000000000000050c3000000000000691871639356419d911f2e0df2f5e015ef5672041d5a5a52c10361646467e827bf96529b5780ad0702757b8bad315e2bb88000014140fb04a7792ffac8d8c777dbf7bce6c016f8d8e732338dbe117bec03d7f6dd1ddf1b508a387aff93c1cf075467c1d0e04b00eb9f3d08976e02758081cc8937f38f232102fb608eb6d1067c2a0186221fab7669a7e99aa374b94a72f3fc000e5f1f5c335eac"
    },
    "error_code": 0,
    "error_info": ""
}
```
signeovminvokeabitx method use the signer account to pay network fee, if you want to use other account to payer the fee, please use payer parameter to specifies。
Note that if specifies payer parameter, don't forget to use sigrawtx method to sign the transaction output by signeovminvokeabitx with the fee payer's account.

Examples:
```
{
    "gas_price":XXX,    //gasprice
    "gas_limit":XXX,    //gaslimit
    "address":"XXX",    //The NeoVM contract address
    "params":[XXX],     //The parameters of the NeoVM contract are constructed according to the ABI of calling method. All values are string type.
    "payer":"XXX",      //The fee payer's account address
    "contract_abi":XXX, //The ABI of contract
}
```

### 2.9 Create Account

Sigsvr can also create account. The account created by sigsvr is ECDSA with 256 bits key pair, and using SHA256withECDSA as signature scheme.

Note than in order wont lose the account created by sigsvr, please backup wallet data on time, and backup sigsvr log.

Method Name: createaccount

Request parameters: null

Reponse:
```
{
    "account":XXX     //The address of account created by sigsvr
}
```

Examples

Request:
```
{
    "qid":"t",
    "method":"createaccount",
    "pwd":"XXXX",     //The unlock password of account created by sigsvr
    "params":{}
}
```

Response:
```
{
    "qid": "t",
    "method": "createaccount",
    "result": {
        "account": "AG9nms6VMc5dGpbCgrutsAVZbpCAtMcB3W"
    },
    "error_code": 0,
    "error_info": ""
}
```

### 2.10 ExportAccount

Export account method can export accounts in wallet data into a wallet file, as backup of accounts

Method Name: exportaccount

Request parameters:

```
{
    "wallet_path": "XXX"    //Save directory path of exported wallet file. If don't specifics, will use sigsvr's current path.
}
```

Response result:
```
{
     "wallet_file": "XXX"   //Full path of exported wallet file.
     "account_num": XXX     //Account number of exported wallet.
}
```


Examples

Request:
```
{
	"qid":"t",
	"method":"exportaccount",
	"params":{}
}
```

Response:
```
{
    "qid": "t",
    "method": "exportaccount",
    "result": {
        "wallet_file": "./wallet_2018_08_03_23_20_12.dat",
        "account_num": 9
    },
    "error_code": 0,
    "error_info": ""
}
```

