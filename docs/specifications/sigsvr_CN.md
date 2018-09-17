# Ontology 签名服务器使用说明

[[English](sigsvr.md)|中文]

Ontology签名服务器sigsvr是一个用于对交易进行签名的rpc服务器。

* [Ontology 签名服务器使用说明](#ontology-签名服务器使用说明)
	* [1、签名服务启动](#1-签名服务启动)
		* [1.1 签名服务启动参数：](#11-签名服务启动参数)
		* [1.2 导入钱包账户](#12-导入钱包账户)
			* [1.2.1 导入钱包账户参数](#121-导入钱包账户参数)
		* [1.3 启动](#13-启动)
	* [2、签名服务方法](#2-签名服务方法)
		* [2.1 签名服务调用方法](#21-签名服务调用方法)
		* [2.2 对数据签名](#22-对数据签名)
		* [2.3 对普通交易签名](#23-对普通交易签名)
		* [2.4 对普通方法多重签名](#24-对普通方法多重签名)
		* [2.5 转账交易签名](#25-转账交易签名)
		* [2.6 Native合约调用签名](#26-native合约调用签名)
			* [举例1: 构造普通转账交易](#举例1-构造普通转账交易)
			* [举例2: 构造提取ONG交易](#举例2-构造提取ong交易)
		* [2.7 NeoVM合约调用签名](#27-neovm合约调用签名)
		* [2.8 NeoVM合约ABI调用签名](#28-neovm合约abi调用签名)
		* [2.9 创建账户](#29-创建账户)
		* [2.10 导出钱包账户](#210-导出钱包账户)

## 1、签名服务启动

### 1.1 签名服务启动参数：

--loglevel
loglevel 参数用于设置sigsvr输出的日志级别。sigsvr支持从0:Trace 1:Debug 2:Info 3:Warn 4:Error 5:Fatal 6:MaxLevel 的7级日志，日志等级由低到高，输出的日志量由多到少。默认值是2，即只输出info级及其之上级别的日志。

--walletdir
walletdir 参数用于设置钱包数据存储目录。默认值为:"./wallet_data"。

--cliaddress
cliaddress 参数用于指定sigsvr启动时绑定的地址。默认值为127.0.0.1，仅接受本机的请求。如果需要被网络中的其他机器访问，可以指定网卡地址，或者0.0.0.0。

--cliport
签名服务器绑定的端口号。默认值为20000。

--abi
abi 参数用于指定签名服务所使用的native合约abi目录，默认值为./abi

### 1.2 导入钱包账户

签名服务在启动前，应该先导入钱包账户。

#### 1.2.1 导入钱包账户参数

--walletdir
walletdir 参数用于设置钱包数据存储目录。默认值为:"./wallet_data"。

--wallet
待导入钱包路径。默认值为："./wallet.dat"。

**导入钱包账户命令**

```
./sigsvr import
```

### 1.3 启动

```
./sigsvr
```

## 2、签名服务方法

签名服务目前支持对数据签名，对普通交易的签名和多重签名，构造ONT/ONG转账交易并对交易签名，构造Native合约调用交易并对交易签名，构造NeoVM合约调用交易并对交易签名。

### 2.1 签名服务调用方法

签名服务是一个json rpc服务器，采用POST方法，请求的服务路径统一为：

```
http://localhost:20000/cli
```
请求结构：

```
{
    "qid":"XXX",    //请求ID，同一个应答会带上相同的qid
    "method":"XXX", //请求的方法名
    "account":"XXX",//签名账户
    "pwd":"XXX",    //账户解锁密码
    "params":{
    	//具体方法的请求参数,按照调用的请求方法要求填写
    }
}
```
应答结构：

```
{
    "qid": "XXX",   //请求ID
    "method": "XXX",//请求的方法名
    "result": {     //应答结果
        "signed_tx": "XXX"  //签名后的交易
    },
    "error_code": 0,//错误码，0表示成功，非0表示失败
    "error_info": ""//错误描述
}
```

错误码：

错误码  | 错误说明
--------|-----------
1001 | 无效的http方法
1002 | 无效的http请求
1003 | 无效的请求参数
1004 | 不支持的方法
1005 | 账户未解锁
1006 | 无效的交易
1007 | 找不到ABI
1008 | ABI不匹配
9999 | 未知错误

### 2.2 对数据签名

可以对任意数据签名，需要注意的是带签名的数据需要用16进制编码成字符串。

方法名：sigdata

请求参数：

```
{
    "raw_data":"XXX"    //待签名的数据（用16进制编码后的数据）
}
```
应答结果：

```
{
    "signed_data": "XXX"//签名后的数据（用16进制编码后的数据）
}
```

举例：

请求：

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
应答：

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

### 2.3 对普通交易签名

方法名：sigrawtx

请求参数：

```
{
    "raw_tx":"XXX"      //待签名的交易
}
```
应答结果：

```
{
    "signed_tx":"XXX"   //签名后的交易
}
```
举例

请求：
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
应答：

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
### 2.4 对普通方法多重签名
多签签名由于私钥掌握在不同的人手上，应该多重签名方法需要被多次调用。
方法名：sigmutilrawtx
请求参数：
```
{
    "raw_tx":"XXX", //待签名的交易
    "m":xxx         //多重签名中，最少需要的签名数
    "pub_keys":[
        //签名的公钥列表
    ]
}
```
应答结果：

```
{
    "signed_tx":"XXX" //签名后的交易
}
```
举例
请求：

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
应答：

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
### 2.5 转账交易签名
为了简化转账交易签名过程，提供了转账交易构造功能，调用的时候，只需要提供转账参数即可。

方法名称：sigtransfertx
请求参数：
```
{
	"gas_price":XXX,  //gasprice
	"gas_limit":XXX,  //gaslimit
	"asset":"ont",    //asset: ont or ong
	"from":"XXX",     //付款账户
	"to":"XXX",       //收款地址
	"amount":"XXX"      //转账金额。注意，由于ong的精度是9，应该在进行ong转账时，需要在实际的转账金额上乘以1000000000。
}
```
应答结果：

```
{
    "signed_tx":XXX     //签名后的交易
}
```

举例
请求：

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

应答：

```
{
    "qid": "t",
    "method": "sigtransfertx",
    "result": {
        "signed_tx": "00d177b8315b00000000000000003075000000000000084c8f4060607444fc95033bd0a9046976d3a9f57100c66b147ce25d11becca9aa8f157e24e2a14fe100db73466a7cc814fc8a60f9a7ab04241a983817b04de95a8b2d4fb86a7cc85a6a7cc86c51c1087472616e736665721400000000000000000000000000000000000000010068164f6e746f6c6f67792e4e61746976652e496e766f6b6500014140115fbc5ab7e2260ea5c7093f7fd93bdb0e874654e5b04b9b1539f1a40fde9f459a5844dfd280df02e3092e6967bfa88025456ec637ac4bbe20b1a3cf71be98262321035f363567ff82be6f70ece8e16378871128788d5a067267e1ec119eedc408ea58ac"
    },
    "error_code": 0,
    "error_info": ""
}
```

sigtransfertx方法默认使用签名账户作为手续费支付方，如果需要使用其他账户作为手续费的付费账户，可以使用payer参数指定。
注意：如果指定了手续费付费账户，还需要调用sigrawtx方法，使用手续费账户对sigtransfertx方法生成的交易进行签名，否则会导致交易执行失败。

举例
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
    	"amount":"10",
    	"payer":"ARVVxBPGySL56CvSSWfjRVVyZYpNZ7zp48"
    }
}
```


### 2.6 Native合约调用签名

Native合约调用交易根据ABI构造，并签名。

注意：
sigsvr启动时，默认会在当前目录下查找"./abi"下的native合约abi。如果naitve目录下没有该合约的abi，会返回1007错误。Native 合约abi的查询路径可以通过--abi参数设定。

方法名称：signativeinvoketx
请求参数：

```
{
    "gas_price":XXX,    //gasprice
    "gas_limit":XXX,    //gaslimit
    "address":"XXX",    //调用native合约的地址
    "method":"XXX",     //调用native合约的方法
    "version":0,        //调用native合约的版本号
    "params":[
        //具体合约 Native合约调用的参数根据调用方法的ABI构造。所有值都使用字符串类型。
    ]
}
```
应答结果：
```
{
    "signed_tx":XXX     //签名后的交易
}
```

以构造ont转账交易举例
请求：

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
应答：

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

signativeinvoketx 方法默认使用签名账户作为手续费支付方，如果需要使用其他账户作为手续费的付费账户，可以使用payer参数指定。
注意：如果指定了手续费付费账户，还需要调用sigrawtx方法，使用手续费账户对 signativeinvoketx 方法生成的交易进行签名，否则会导致交易执行失败。

#### 举例1: 构造普通转账交易
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
    	"payer":"ARVVxBPGySL56CvSSWfjRVVyZYpNZ7zp48",
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

#### 举例2: 构造提取ONG交易

``` json
{
	"Qid":"t",
	"Method":"signativeinvoketx",
	"account":"ARVVxBPGySL56CvSSWfjRVVyZYpNZ7zp48",	  //提取账户
	"pwd":"XXX",
	"Params":{
		"gas_price":5000,
		"gas_limit":20000,
		"address":"0200000000000000000000000000000000000000",
		"method":"transferFrom",
		"version":0,
		"params":[
			"ARVVxBPGySL56CvSSWfjRVVyZYpNZ7zp48",	//提取账户
			"AFmseVrdL9f9oyCzZefL9tG6UbvhUMqNMV",   //ONT合约地址(base58格式)
			"ARVVxBPGySL56CvSSWfjRVVyZYpNZ7zp48",  //ONG接受地址，可以于提取地址不一样
			"310860000000000"												//提取金额(需要在实际金额上乘以10的9次方)
		]
	}
}
```

### 2.7 NeoVM合约调用签名

NeoVM合约调用根据要调用的NeoVM合约构造调用交易，并签名。

NeoVM参数合约支持array、bytearray、string、int以及bool类型，构造参数时需要提供参数类型及参数值，参数值统一使用字符串类型。array是对象数组，数组元素支持任意NeoVM支持的参数类型和数量。

方法名称：signeovminvoketx
请求参数：
```
{
    "gas_price":XXX,    //gasprice
    "gas_limit":XXX,    //gaslimit
    "address":"XXX",    //调用Neovm合约的地址
    "params":[
        //具体合约 Neovm合约调用的参数，根据需要调用的具体合约构造。所有值都使用字符串类型。
    ]
}
```
应答结果：
```
{
    "signed_tx":XXX     //签名后的交易
}
```

举例
请求:

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
应答：

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

signeovminvoketx 方法默认使用签名账户作为手续费支付方，如果需要使用其他账户作为手续费的付费账户，可以使用payer参数指定。
注意：如果指定了手续费付费账户，还需要调用sigrawtx方法，使用手续费账户对 signeovminvoketx 方法生成的交易进行签名，否则会导致交易执行失败。

举例
```
{
    "gas_price":XXX,    //gasprice
    "gas_limit":XXX,    //gaslimit
    "address":"XXX",    //调用Neovm合约的地址
    "payer":"XXX",      //手续费付费地址
    "params":[
        //具体合约 Neovm合约调用的参数，根据需要调用的具体合约构造。所有值都使用字符串类型。
    ]
}
```

### 2.8 NeoVM合约ABI调用签名

NeoVM合约ABI调用签名，需要提供合约的abi，以及合约调用的参数，其中所有的参数都是字符串类型。

方法名称：signeovminvokeabitx

请求参数：

```
{
    "gas_price":XXX,    //gasprice
    "gas_limit":XXX,    //gaslimit
    "address":"XXX",    //调用Neovm合约的地址
    "params":[XXX],     //调用参数（所有的参数都是字符串类型）
    "contract_abi":XXX, //合约ABI
}
```
应答

```
{
    "signed_tx":XXX     //签名后的交易
}
```
举例

请求：

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
应答:

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
signeovminvokeabitx 方法默认使用签名账户作为手续费支付方，如果需要使用其他账户作为手续费的付费账户，可以使用payer参数指定。
注意：如果指定了手续费付费账户，还需要调用sigrawtx方法，使用手续费账户对 signeovminvokeabitx 方法生成的交易进行签名，否则会导致交易执行失败。

举例
```
{
    "gas_price":XXX,    //gasprice
    "gas_limit":XXX,    //gaslimit
    "address":"XXX",    //调用Neovm合约的地址
    "params":[XXX],     //调用参数（所有的参数都是字符串类型）
    "payer":"XXX",      //手续费付费地址
    "contract_abi":XXX, //合约ABI
}
```

### 2.9 创建账户

可以通过签名服务创建新账户。新创建的账户使用256位的ECDSA密钥，并使用SHA256withECDSA作为签名模式。

注意：使用签名服务创建密码后，为了安全起见，需要及时备份账户密钥，以防丢失；同时保留好备份后的签名服务运行日志，日志中记录最新创建的密钥信息。

方法名称：createaccount

请求参数：无

应答

```
{
    "account":XXX     //新创建的账户地址
}
```
举例
请求：

```
{
	"qid":"t",
	"method":"createaccount",
	"pwd":"XXXX",     //新创建账户的解锁密码
	"params":{}
}
```
应答：

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

### 2.10 导出钱包账户

导出钱包账户命令可以把钱包数据中的账户导出到一个钱包文件，作为账户密钥备份使用。

方法名称：exportaccount

请求参数:

```
{
    "wallet_path":"XXX" //导出钱包文件存储目录, 如果填，则默认存储于签名服务运行时的当前目录下。
}
```

应答:

```
{
   "wallet_file": "XXX",//导出的钱包文件路径及名称
   "account_num": XXX   //导出的账户数量
}
```

举例

请求：

```
{
	"qid":"t",
	"method":"exportaccount",
	"params":{}
}
```

应答：
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

