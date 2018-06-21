# Ontology 签名服务器使用说明

Ontology签名服务器sigsvr是一个用于对交易进行签名的rpc服务器。签名服务器绑定在127.0.0.1地址上，只支持本机发送的签名请求。

## 1、签名服务启动

### 1.1 签名服务启动参数：

--loglevel
loglevel 参数用于设置sigsvr输出的日志级别。sigsvr支持从0:Debug 1:Info 2:Warn 3:Error 4:Fatal 5:Trace 6:MaxLevel 的7级日志，日志等级由低到高，输出的日志量由多到少。默认值是1，即只输出info级及其之上级别的日志。

--wallet, -w
wallet 参数用于指定sigsvr启动时的钱包文件路径。默认值为"./wallet.dat"。

--account, -a
account 参数用于指定sigsvr启动时加载的账户地址。不填则使用钱包默认账户。

--cliport
签名服务器绑定的端口号。默认值为20000。

### 1.2 启动

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
	"amount":XXX      //转账金额。注意，由于ong的精度是9，应该在进行ong转账时，需要在实际的转账金额上乘以1000000000。
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
	"params":{
		"gas_price":0,
		"gas_limit":30000,
		"asset":"ont",
		"from":"TADPWFL3yHoHeNN3Bx1RUCrrXjU35EmeXp",
		"to":"TA5gYXCSiUq9ejGCa54M3yoj9kfMv3ir4j",
		"amount":10
	}
}
```

应答：

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

### 2.6 Native合约调用签名

Native合约调用交易根据ABI构造，并签名。

注意：
sigsvr启动时，会在当前目录下查找"./cmd/abi/native"下的native合约abi。如果naitve目录下没有该合约的abi，怎会返回1007错误。

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
	"Params":{
		"gas_price":0,
		"gas_limit":50000,
		"address":"ff00000000000000000000000000000000000001",
		"method":"transfer",
		"version":0,
		"params":[
			[
				[
				"TA587BCw7HFwuUuzY1wg2HXCN7cHBPaXSe",
				"TA5gYXCSiUq9ejGCa54M3yoj9kfMv3ir4j",
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
    "qid": "1",
    "method": "signativeinvoketx",
    "result": {
        "signed_tx": "00d1a55b175b00000000000000003075000000000000011e68f7bf0aaba1f18213639591f932556eb674ff4c0000ff00000000000000000000000000000000000001087472616e736665722c01011e68f7bf0aaba1f18213639591f932556eb67401397cafa8cc71ae9e555e439fc0f0a5ded12a2afde803000101231202026940ba3dba0a385c44e4a187af75a34e281b96200430db2cbc688a907e5fb545010141013eaf77601c847c4244b5b5a1b22153d6231c29853838b8a963c34834987f86cb6f63a8e5b454899457ee0be56eec16339efdbc4876649c72f313f04c5fa2ba11"
    },
    "error_code": 0,
    "error_info": ""
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

