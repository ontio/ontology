[English](./ontid.md) / 中文

<h1 align="center">本体分布式身份标识</h1>
<p align="center" class="version">Contract Version 1.0</p>


实体是指现实世界中的个人、组织（组织机构、企事业单位等）、物品（手机、汽车、IOT设备等）、内容（文章、版权等），而身份是指实体在网络上的对应标识。本体使用本体⾝份标识（ONT ID）来标识和管理实体的网络身份。在本体上，⼀个实体可以对应到多个身份标识，且多个身份标识之间没有任何关联。

ONT ID是⼀个去中心化的身份标识协议，ONT ID具有去中心化、自主管理、隐私保护、安全易用等特点。每⼀个ONT ID都会对应到⼀个ONT ID描述对象（ONT DDO）。

> ONT ID协议已被本体区块链智能合约完整实现。作为协议层，它遵循解耦设计，并不仅限于本体区块链，同样可以基于其他区块链。

## 1. 身份标识协议规范

### 1.1 ONT ID 生成

ONT ID是一种URI，由每个实体自己生成成，生成算法需要保证碰撞概率非常低。同时在向本体注册时，共识节点会检查该ID是否已被注册。

ONT ID 生成算法：

为了防止用户错误输入ONT ID，我们定义一个合法的ONT ID必须包含4个字节的校验数据。下面详细描述一下如何生成一个合法的ONT ID。

 1. 生成32字节临时随机数nonce，计算h = Hash160(nonce），data = VER || h；
 2. 计算出data的一个4字节校验，即checksum = SHA256(SHA256(data))[0:3]；
 3. 令idString = data || checksum；
 4. 将"did:ont:"与data级联，即 ontId = "did:ont:" || idString；
 5. 输出ontId。

如上所述，`<ont>`是网络标识符，`<VER>`是1个字节的版本标签。 在ONT中，`<VER> = 41，<ont> =“ont”`。 也就是说，ONTology中身份的前8个字节是“did：ont：”，外加一个25字节长的idString，构成一个完整的ONT ID。

### 1.2 自主管理

本体利用数字签名技术保障实体对自己身份标识的管理权。ONT ID在注册时即与实体的公钥绑定，从而表明其所有权。对ONT ID的使用及其属性的修改需要提供所有者的数字签名。实体可以自主决定ONT ID的使用范围，设置ONT ID绑定的公钥，以及管理ONT ID的属性。

### 1.3 多密钥绑定

本体支持多种国内、国际标准化的数字签名算法，如ECDSA、SM2等。ONT ID绑定的密钥需指定所使用的算法，同时一个ONT ID可以绑定多个不同的密钥，以满足实体在不同的应用场景的使用需求。

### 1.4 身份丢失恢复

ONT ID的所有者可以设置恢复人代替本人行使对ONT ID的管理权，如修改ONT ID对应的属性信息，在密钥丢失时替换密钥。恢复人可以实现多种访问控制逻辑，如“与”、“或”、“(m, n)-门限”。请参阅 [附录 B](#b.-recovery-account-address) 获取更多细节。

### 1.5 身份描述对象DDO规范

ONT ID对应的身份描述对象DDO存储在本体区块链，由DDO的控制人写入到区块链，并向所有用户开放读取。

DDO规范包含如下信息：
- 公钥列表`PublicKeys`：用户用于身份认证的公钥信息，包括公钥id、公钥类型、公钥数据；
- 属性对象`Attributes`：所有的属性构成一个JSON对象；
- 恢复人地址`Recovery`：恢复人可帮助重置用户公钥列表。

以下是一个json格式的DDO样例,

```json
{
	"ONTId": "did:ont:TVuF6FH1PskzWJAFhWAFg17NSitMDEBNoa",
	"Owners": [{
			"PubKeyId": "did:ont:TVuF6FH1PskzWJAFhWAFg17NSitMDEBNoa#keys-1",
			"Type": "ECDSA",
			"Curve": "nistp256",
			"Value":"022f71daef10803ece19f96b2cdb348d22bf7871c178b41f35a4f3772a8359b7d2"
		}, {
			"PublicKeyId": "did:ont:TVuF6FH1PskzWJAFhWAFg17NSitMDEBNoa#keys-2", 
			"Type": "RSA",
			"Length": 2048, 
			"Value": "3082010a...."
		}
	],
	"Attributes": {
		"OfficialCredential": {
			"Service": "PKI",
			"CN": "ont.io",
			"CertFingerprint": "1028e8f7043f12c0c2069bd7c7b3b26213964566"
		}
	},
	"Recovery": "TA63T1gxXPXWsBqHtBKcV4NhFBhw3rtkAF"
}
```

## 2. 智能合约实现说明

ONT ID协议在Ontology Blockchain平台上以原生合约的形式实现。通过ONT ID合约，用户可以注册ID，管理自己的公钥列表，修改个人的信息，并添加账户恢复人。

对于大多数方法，如果执行成功，将会推送一条事件消息来通知调用者。 请参阅 [事件](#2.2.-events) 小节.

## 2.1 调用方法

**注意事项**: 用作以下方法参数的用户公钥应该是构造交易的公钥，即它可以通过交易的签名验证，且该公钥应该已绑定到用户的ID（注册方法除外）。

###  a. 身份登记

用户在登记身份时，必须要提交一个公钥，并且这次操作必须是由该公钥发起。

```
方法: regIDWithPublicKey

参数:
    0    byte array    用户ID
    1    byte array    公钥

事件: 当成功时触发'Register'事件
```
 
### b. 增加控制密钥

用户在自己的公钥列表中添加一个新公钥。

```
方法: addKey

参数:
    0    byte array    用户ID
    1    byte array    欲添加的新公钥
    2    byte array    用户公钥或恢复地址

事件: 当成功时触发'PublicKey'事件
```

### c. 删减控制密钥

从用户的公钥列表中，移除一个公钥。

```
方法: removeKey

参数:
    0    byte array    用户ID
    1    byte array    欲删除的公钥
    2    byte array    用户公钥或恢复地址

事件: 当成功时触发'PublicKey'事件
```
	
### d. 密钥恢复机制

添加与修改账户的恢复人。

```
方法: addRecovery

参数:
    0    byte array    用户ID
    1    byte array    恢复地址
    2    byte array    用户公钥

事件: 当成功时触发'Recovery'事件
```

当且仅当参数2是用户现有的公钥并且恢复地址尚未设置时，此方法才能成功。


```
Method: changeRecovery

Arguments:
    0    byte array    用户ID
    1    byte array    新的恢复地址
    2    byte array    原先的恢复地址

事件: 当成功时触发'Recovery'事件
```

这次合约调用必须是由原先的恢复地址发起。

### e. 属性管理

一个属性包含以下3个参数:

```
attribute {
    key     // byte array
    type    // byte array
    value   // byte array
}
```

当注册ID的同时, 用户可以用 `regIDWithAttributes`设置属性.

```
方法: regIDWithAttributes

参数:
    0    byte array    用户ID
    1    byte array    用户公钥
    2    byte array    序列化的属性
```

按以下步骤生成参数2:

1. 对于每个属性, 对3个参数依次序列化.
2. 对序列化后的属性进行拼接.

用户个人属性的增删改，均需要得到用户的授权。

```
方法: addAttributes

参数:
    0    byte array    用户ID
    1    byte array    序列化的属性
    2    byte array    用户公钥

事件: 当成功时触发'Attribute'事件
```

参数1与regIDWithAttributes的参数2相同.

如果一个属性不存在，该属性将被添加。如果存在，原始属性将被更新。

```
方法: removeAttribute

参数:
    0    byte array    用户ID
    1    byte array    欲删除属性的密钥
    2    byte array    用户公钥

事件: 当成功时触发'Attribute'事件
```


### f. 查询身份信息

查询方法通过预执行的方式调用，不需要发送交易到网络，直接从本地存储获得结果。

#### 密钥

```
方法: getPublicKeys

参数: byte array, 用户ID

返回结果: byte array
```

每个公钥包含以下属性:

```
publicKey {
    index   // 32 bits unsigned int
    data    // byte array
}
```

返回的字节数组由一系列序列化的公钥组成，即`publicKey array`的序列化.

索引是出现在DDO的PubKeyID中的数字。 它在被注册或被添加时自动生成。被撤销的密钥的索引不会被回收。该密钥的状态会被标识为已撤回。用以下方法可获得密钥状态：

```
方法: getKeyState

参数:
    0    byte array  user's ID
    1    uint32      key index

返回结果: "in use" | "revoked" | "not exist"
```

#### 属性

```
方法: getAttributes

参数: byte array, 用户ID

返回结果: byte array
```

返回的字节数组由一系列序列化后的属性组成，其结构如前文定义。

#### DDO

```
方法: getDDO

参数: byte array, 用户ID

返回结果: byte array
```

返回值包含 `GetPublicKeys` 和 `GetAttributes`的结果和恢复地址:

```
ddo {
    byte[] keys         // publicKey[]的序列化
    byte[] attributes   // attribute[]的序列化
    byte[] recovery     // 恢复地址
}
```


## 2.2. 事件

有三种事件消息：

- `Register`:  会推送与身份登记有关的信息。

	| Field | Type       | Description       |
	| :---- | :--------- | :---------------- |
	|  op   | string     | 消息类型      |
	|  ID   | byte array | 注册的 ONT ID |

- `PublicKey`: 会推送与公钥操作有关的消息。

	| Field      | Type       | Description    |
	| :--------- | :--------- | :--------------|
	|  op        | string     | 消息类型："add" 或 "remove" |
	|  ID        | byte array | 用户的 ONT ID    |
	| key index  | uint32     | 密钥索引 |
	| public key | byte array | 公钥数据  |

- `Attribute`: 会推送与属性操作有关的消息。

	| Field    | Type       | Description    |
	| :------- | :--------- | :------------- | 
	| op       | string     | 消息类型："add"、remove"  |
	| ID       | byte array | 用户的 ONT ID|
	| attrName | byte array | 属性名 |
	
- `Recovery`: Push the messages related to recovery operations.

	| Field    | Type       | Description    |
	| :------- | :--------- | :------------- | 
	| op       | string     | 消息类型："add"、"change" |
	| ID       | byte array | 用户的 ONT ID |
	| address  | byte array | 恢复地址 |

## 附录

### A. 公钥数据

类型字节数组的公钥数据遵循[ontology-crypto serialization](https://github.com/ontio/ontology-crypto/wiki/Asymmetric-Keys#Serialization)中定义的格式:

    algorithm_id || algorithm_parameters || public_key


### B. 恢复账户地址

恢复帐户可以实现各种访问控制逻辑，如（m，n）- 门限控制。 一个（m，n）门限控制账户由n个公钥共同管理。 要使用它，您必须至少收集m个有效签名。

- `(m, n) threshold` 控制账户

	```
	0x02 || RIPEMD160(SHA256(n || m || publicKey_1 || ... || publicKey_n))
	```

- `AND` 控制账户
   
   这相当于（n，n）门限控制账户。

- `OR` 控制账户
  
   这相当于（1，n）门限控制账户。

