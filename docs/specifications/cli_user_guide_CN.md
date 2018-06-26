# Ontology cli 使用说明

Ontology cli 是Ontology命令行客户端，用于启动和管理Ontology节点，管理钱包 账户，发送交易以及部署和调用智能合约等。

## 1、启动和管理Ontology节点

Ontology cli有很多启动参数，用于配置很管理Ontology节点的一些行为。如果不带任何参数启动Ontology cli时，默认会作为一个同步节点的接入Ontology的polaris测试网络。

```
./ontology
```
使用./ontology -help 可以查看到Ontology cli节点支持的所有启动参数。

### 1.1 启动参数

以下是Ontology cli 支持的命令行参数：

#### 1.1.1 Ontology 系统参数

--config
config 参数用于指定当前Ontology节点创世区块配置文件的路径。如果不指定，将使用Polaris测试网的创世块配置。注意，同一个网络所有节点的创世区块配置必须一致，否则会因为区块数据不兼容导致无法启动节点或同步区块数据。

--loglevel
loglevel 参数用于设置Ontology输出的日志级别。Ontology支持从0:Debug 1:Info 2:Warn 3:Error 4:Fatal 5:Trace 6:MaxLevel 的7级日志，日志等级由低到高，输出的日志量由多到少。默认值是1，即只输出info级及其之上级别的日志。

--disableeventlog
disableeventlog 参数用于关闭智能合约执行时输出的event log，以提升节点交易执行性能。Ontology 节点默认会开启智能合约执行时的event log输出功能。

--datadir
datadir 参数用于指定区块数据的存放目录。默认值为"./Chain"。

--import
import 参数用于启动Ontology节点的区块导入功能，通过导入本地文件的方式来提高区块同步速度。

--importheight
importheight 参数配合--import使用，用于指定导入的终止区块高度。如果importheight指定的区块高度小于区块文件的最大高度时，只导入到importheight指定的高度，剩余的区块会停止导入。默认值为0，表示导入所有的区块。

--importfile
importfile 参数配合--import使用，用于区块导入时指定导入文件的路径。默认值为"./blocks.dat"。

#### 1.1.2 账户参数

--wallet, -w
wallet 参数用于指定Ontology节点启动时的钱包文件路径。默认值为"./wallet.dat"。

--account, -a
account 参数用于指定Ontlogy节点启动时的账户地址。不填则使用钱包默认账户。

--password, -p
password 参数用于指定Ontology节点启动的账户密码。因为在命令行中输入的账户密码会被保存在系统的日志中，容易造成密码泄露，因此在生产环境中建议不要使用该参数。

#### 1.1.3 共识参数

--disableconsensus
disableconsensus 参数用于关闭网络共识。在启动同步节点的情况下，建议开启该参数，关闭网络共识。默认是开启网络共识的。

--maxtxinblock
maxtxinblock 参数用于设置区块最大的交易数量。默认值是50000。

#### 1.1.4 P2P网络参数

--networkid
networkid 参数用于指定网络ID，networkid不同将无法连接到区块链网络中。1:主网, 2:polaris测试网络, 3:testmode测试网, 其他的是用户自定义网络。

--nodeport
nodeport 参数用于指定P2P网络端口号，默认值为20338。

--consensusport
consensusport 参数用于指定共识网络端口号。默认情况下，共识网络复用P2P网络，因此不需要指定共识网络端口，在通过--dualport参数启动双网络后，则需要单独设置共识网络端口号。默认值为20339。

--dualport
dualport 参数启动双网络，即用于处理交易消息的P2P网络，和用于共识消息的共识网络。默认不开启。

#### 1.1.5 RPC 服务器参数

--disablerpc
disablerpc 参数用于关闭rpc服务器。Ontology节点在启动时会默认启动rpc服务器。

--rpcport
rpcport 参数用指定rpc服务器绑定的端口号。默认值为20336。

#### 1.1.6 Restful 服务器参数

--rest
rest 参数用于启动rest服务器。

--restport
restport 参数用于指定restful服务器绑定的端口号。默认值为20334。

#### 1.1.7 Web socket服务器参数

--ws
ws 参数用于启动Web socket服务器。

--wsport
wsport 参数用于指定Web socket服务器绑定的端口号。默认值为20335

#### 1.1.8 测试模式参数

--testmode
testmode 参数用于启动单节点的测试网络，便于开发和调试。使用testmode启动测试网络时，会同时启动rpc、rest以及ws服务器，同时会删除上一次使用testmode时创建的区块数据。

--testmodegenblocktime
testmodegenblocktime 参数用于设置测试模式下的出块时间，时间单位为秒，最小出块时间为2秒，默认值为6秒。

#### 1.1.9 交易参数

--gasprice
gasprice 参数用于设定当前节点交易池接受交易的最低gasprice，低于这个gasprice的交易将会被丢弃。在交易池有交易排队等待打包进区块时，交易池根据gas price的高低来排序交易，gas price高的交易将会被优先处理。默认值为0。

--gaslimit
gaslimit 参数用于设置当前节点交易池接受交易的最低gaslimit，低于这个gaslimit的交易将被丢弃。默认值为30000。


### 1.2 节点部署

#### 1.2.1 创世区块配置文件

Ontology支持VBFT和dBFT共识算法，一个区块链网络必须使用同一种共识算法，否则节点将无法正常运行。Ontology内置了Polaris测试网的配置，如果启动Ontology时不指定创世块配置文件，将会使用Polaris测试网络的创世块配置；如果需要搭建Ontology私有链，则需要通过--config指定创世块配置文件的路径。

使用dBFT共识算法部署Ontology测试网络最少需要4个节点，使用VBFT共识算法部署Ontology测试网络最少需要7个节点。

##### 1.2.1.1 VBFT配置文件

使用VBFT共识算法的创世块配置文件示例：

```json
{
  "SeedList": [
    "192.168.0.1:20338"             //种子节点列表
  ],
  "ConsensusType":"vbft",           //指定当前网络使用VBFT算法
  "VBFT":{                          //VBFT算法配置
    "n":7,                          //网络规模（暂时未使用）
    "c":2,                          //容错节点数量
    "k":7,                          //共识节点数量
    "l":112,                        //POS表长度
    "block_msg_delay":10000,        //区块消息最大广播延迟(ms)
    "hash_msg_delay":10000,         //哈希消息最大广播延迟(ms)
    "peer_handshake_timeout":10,    //节点握手超时时间(s)
    "max_block_change_view":1000,   //共识周期
    "vrf_value":"1c9810aa9822e511d5804a9c4db9dd08497c31087b0daafa34d768a3253441fa20515e2f30f81741102af0ca3cefc4818fef16adb825fbaa8cad78647f3afb590e",
    "vrf_proof":"c57741f934042cb8d8b087b44b161db56fc3ffd4ffb675d36cd09f83935be853d8729f3f5298d12d6fd28d45dde515a4b9d7f67682d182ba5118abf451ff1988",
    "peers":[                       //记账节点配置
      {
        "index":1,                  //记账节点索引
        "peerPubkey":"1202028541d32f3b09180b00affe67a40516846c16663ccb916fd2db8106619f087527",  //记账节点公钥
        "address":"TA9TVuR4Ynn4VotfpExY5SaEy8a99obFPr",                                         //记账节点账户地址
        "initPos":1000              //初始ONT抵押值
      },
      {
        "index":2,
        "peerPubkey":"120202dfb161f757921898ec2e30e3618d5c6646d993153b89312bac36d7688912c0ce",
        "address":"TA9TVuR4Ynn4VotfpExY5SaEy8a99obFPr",
        "initPos":2000
      },
      {
        "index":3,
        "peerPubkey":"1202039dab38326268fe82fb7967fe2e7f5f6eaced6ec711148a66fbb8480c321c19dd",
        "address":"TA9TVuR4Ynn4VotfpExY5SaEy8a99obFPr",
        "initPos":3000
      },
      {
        "index":4,
        "peerPubkey":"12020384f2729bc5d9b14dcbf17aba108261dc7ad867127e413d3c8bfb4731739687b3",
        "address":"TA9TVuR4Ynn4VotfpExY5SaEy8a99obFPr",
        "initPos":4000
      },
      {
        "index":5,
        "peerPubkey":"120203362f99284daa9f581fab596516f75475fc61a5f80de0e268a68430dc7589859c",
        "address":"TA9TVuR4Ynn4VotfpExY5SaEy8a99obFPr",
        "initPos":3000
      },
      {
        "index":6,
        "peerPubkey":"120203db6e37a2d897f2d61b42dcd478323a8a20c3444af4ee29653849f38d0bdb67f4",
        "address":"TA9TVuR4Ynn4VotfpExY5SaEy8a99obFPr",
        "initPos":2000
      },
      {
        "index":7,
        "peerPubkey":"12020298fe9f22e9df64f6bfcc1c2a14418846cffdbbf510d261bbc3fa6d47073df9a2",
        "address":"TA9TVuR4Ynn4VotfpExY5SaEy8a99obFPr",
        "initPos":1000
      }
    ]
  }
}
```

##### 1.2.1.2 dBFT配置文件

使用dBFT共识算法的创世块配置文件示例：

```json
{
  "SeedList": [
    "192.168.0.1:20338"               //种子节点列表
  ],
  "ConsensusType":"dbft",             //指定当前网络使用dBFT算法
  "DBFT":{                            //dBFT共识算法配置
    "Bookkeepers": [                  //记账节点公钥列表
      "1202028541d32f3b09180b00affe67a40516846c16663ccb916fd2db8106619f087527",
      "120202dfb161f757921898ec2e30e3618d5c6646d993153b89312bac36d7688912c0ce",
      "1202039dab38326268fe82fb7967fe2e7f5f6eaced6ec711148a66fbb8480c321c19dd",
      "12020384f2729bc5d9b14dcbf17aba108261dc7ad867127e413d3c8bfb4731739687b3"
    ],
    "GenBlockTime":6                  //出块时间间隔，单位是秒
  }
}
```

#### 1.2.2 记账节点部署

按照角色不同，节点可以分为记账节点和同步节点，记账节点参与网络共识，而同步节点只同步记账节点生成的区块。Ontology节点默认会启动Rpc服务器，同时会输出智能合约输出的Event Log，因此如果没有特殊要求，可以使用--disablerpc和--disableeventlog命令行参数关闭rpc和eventlog模块。

推荐记账节点启动参数：

```
./ontology --disablerpc --disableeventlog
```
如果节点没有使用默认的创世块配置文件和钱包账户，可以通过--config参数和--wallet、--account参数指定。
同时，如果记账节点需要修改交易池默认的最低gas price和gas limit，可以通过--gasprice和--gaslimit参数来设定。

#### 1.2.3 同步节点部署

由于同步节点只同步记账节点生成的区块，并不参与网络共识，因此可以通过--disableconsensus参数关闭网络共识模块。

推荐同步节点启动参数：

```
./ontology --disableconsensus
```
如果节点没有使用默认的创世块配置文件和钱包账户，可以通过--config参数和--wallet、--account参数指定。

#### 1.2.4 在单机上部署多节点测试网络

可以在同一台机器上部署Ontology的测试网络，本质上与在多台机器上部署多节点没有什么不同，但是注意使用不同的端口号：

```
./ontology --nodeport=XXX --rpcport=XXX
```
#### 1.2.5 单节点测试网络部署

Ontology支持单节点网络部署，用于开发测试环境搭建。启动单节点测试网络只需要加上--testmode参数即可。

```
./ontology --testmode
```
如果节点没有使用默认的创世块配置文件和钱包账户，可以通过--config参数和--wallet、--account参数指定。
同时，如果记账节点需要修改交易池默认的最低gas price和gas limit，可以通过--gasprice和--gaslimit参数来设定。

## 2、钱包管理

钱包管理命令可以用来添加、查看、修改、删除、导入账户等功能。
使用 ./ontology account --help 命令可以查看钱包管理命令的帮助信息。

### 2.1、添加账户

Ontology支持多种加密算法，包括ECDSA、SM2以及ED25519。

在使用ECDSA加密算法时，可以支持多种密钥曲线，如：P-224、P-256、P-384、P-521；此外，在使用ECDSA加密算法时，还可以指定该密钥的签名方案，如：SHA224withECDSA、SHA256withECDSA、SHA384withECDSA、SHA512withEdDSA、SHA3-224withECDSA、SHA3-256withECDSA、SHA3-384withECDSA、SHA3-512withECDSA、RIPEMD160withECDSA。

在使用SM2加密算法时，使用sm2p256v1曲线，同时使用SM3withSM2签名算法。

使用ED25519加密算法时，使用25519曲线，使用SHA512withEdDSA签名算法。

**默认账户**

每个钱包都一个默认账户，一般情况下是第一个添加的账户。默认账户不能被删除，可以通过./ontology account set 命令来修改默认账户。

#### 2.1.1添加账户参数

--type,t
type参数用于设定加密算法，支持ecdsa, sm2和ed25519加密算法。

--bit-length,b
bit-length参数用于指定密钥长度，如果是ecdsa加密算法，可以选择p-224, p-256, p-384, p-521；如果是sm2加密算法，默认为sm2p256v1；如果是ed25519加密算法，默认为25519。

--signature-scheme,s
signature-scheme参数用于指定密钥签名方案，对于ecdsa加密算法，支持SHA224withECDSA、SHA256withECDSA、SHA384withECDSA、SHA512withEdDSA、SHA3-224withECDSA、SHA3-256withECDSA、SHA3-384withECDSA、SHA3-512withECDSA、RIPEMD160withECDSA这些签名方案；如果是sm2加密算法，默认使用SM3withSM2签名方案；如果使用的是ed25519加密算法，默认使用的是SHA512withEdDSA签名方案。

--default
default参数使用系统默认的密钥方案，默认的密钥方式将会使用ECDSA加密算法，使用P-256曲线以及SHA256withECDSA作为签名算法。

--label
label用于给新创建的账户设置标签，用于方便、快捷查找账户。注意，同一个钱包文件下，不能出现重复的label名。没有设置label的账户则为空字符串。

--wallet
wallet 参数用于指定钱包文件路径。如果钱包文件不存在，则会自动创建一个新的钱包文件。

--number
number参数用于需要创建的账户数量。可以通过number来批量创建账户。number默认值为1。

**添加账户**

```
./ontology account add --default
```

通过 ./ontology account add --help 可以查看帮助信息。

### 2.2、查看账户

使用命令：

```
./ontology account list
```
可以查看当前钱包中的所有账户信息。比如：

```
$ ./ontology account list
Index:1    Address:TA587BCw7HFwuUuzY1wg2HXCN7cHBPaXSe  Label: (default)
Index:2    Address:TA5gYXCSiUq9ejGCa54M3yoj9kfMv3ir4j  Label:
```
其中，Index 为账户在钱包中的索引，索引从1开始，Addres 为账户地址，Label 为账户的标签，default表示当前账户是默认账户。
在Ontology cli中，可以通过Index、Address或非空的Label来查找账户。

使用--v 可以查看账户的详细信息。
通过 ./ontology account list --help 可以查看帮助信息。

### 2.3 修改账户

使用修改账户命令可以修改账户的标签，重新设置默认账户，修改账户密码，如果账户是ECDSA加密算法的密钥，还可以修改密钥的签名方案。
通过 ./ontology account set --help 可以查看帮助信息。

#### 2.3.1修改账户参数

--as-default, -d
as-default参数设置账户为默认账户。

--wallet, -w
wallet参数指定当前操作的钱包路径，默认值为"./wallet.dat"。

--label, -l
label参数用于给账户设置新的标签。注意一个钱包文件中，不能有两个相同的lable。

--changepasswd
changepasswd参数用于修改账户密码。

--signature-scheme, -s
signature-scheme参数用于修改账户签名方案。如果账户使用的是ECDSA密钥，则可以修改如下ECDSA支持的签名方案：SHA224withECDSA、SHA256withECDSA、SHA384withECDSA、SHA512withEdDSA、SHA3-224withECDSA、SHA3-256withECDSA、SHA3-384withECDSA、SHA3-512withECDSA、RIPEMD160withECDSA。

**设置默认账户**

```
./ontology account set --d <address|index|label>
```
**修改账户标签**

```
./ontology account set --label=XXX <address|index|label>
```
**修改账户密码**

```
./ontology account set --changepasswd <address|index|label>
```

**修改ECDSA密钥签名方案**

```
./ontology account set --s=SHA256withECDSA <address|index|label>
```
### 2.4 删除账户

对于钱包中不需要的账户，可以删除。删除账户后无法恢复，所以请谨慎操作。注意：默认账户无法被删除。

```
/ontology account del <address|index|label>
```
### 2.5 导入账户

导入账户命令可以把另一个钱包中的账户导入到当前的钱包中。

#### 2.5.1导入账户参数

--wallet,w
wallet参数指定当前钱包路径，用于接收导入钱包的账户。

--source,s
source参数指定被导入的钱包路径

```
./ontology account import -s=./source_wallet.dat
```

## 3、资产管理

资产管理命令可以查看账户的余额，执行ONT/ONG转账，提取ONG以及查看未绑定的ONG等操作。

### 3.1 查看账户余额

```
./ontology asset balance <address|index|label>
```
### 3.2 ONT/ONG转账

#### 3.2.1转账参数

--wallet, -w
wallet指定转出账户钱包路径，默认值为:"./wallet.dat"

--gasprice
gasprice参数指定转账交易的gas price。交易的gas price不能小于接收节点交易池设置的最低gas price，否则交易会被拒绝。默认值为0。当交易池中有交易在排队等待打包进区块时，交易池会按照gas price由高到低排序，gas price高的交易会被优先处理。

--gaslimit
gaslimit参数指定转账交易的gas limit。交易的gas limit不能小于接收节点交易池设置的最低gas limit，否则交易会被拒绝。gasprice * gaslimit 为账户实际支付的ONG 费用。 默认值为30000。

--asset
asset参数指定转账的资产类型，ont表示ONT，ong表示ONG。默认值为ont。

--from
from参数指定转出账户地址。

--to
to参数指定转入账户地址。

--amount
amount参数指定转账金额。注意：由于ONT的精度是1，因此如果输入的是个浮点值，那么小数部分的值会被丢弃；ONG的精度为9，因此超出9位的小数部分将会被丢弃。

**转账**

```
./ontology asset transfer --from=<address|index|label> --to=<address|index|label> --amount=XXX --asset=ont
```

### 3.3 授权转账

用户可以授权其他账户在授权额度内从本账户中转账。

#### 3.3.1 授权转账参数
--wallet, -w
wallet指定授权转出账户钱包路径，默认值为:"./wallet.dat"

--gasprice
gasprice参数指定转账交易的gas price。交易的gas price不能小于接收节点交易池设置的最低gas price，否则交易会被拒绝。默认值为0。当交易池中有交易在排队等待打包进区块时，交易池会按照gas price由高到低排序，gas price高的交易会被优先处理。

--gaslimit
gaslimit参数指定转账交易的gas limit。交易的gas limit不能小于接收节点交易池设置的最低gas limit，否则交易会被拒绝。gasprice * gaslimit 为账户实际支付的ONG 费用。 默认值为30000。

--asset
asset参数指定转账的资产类型，ont表示ONT，ong表示ONG。默认值为ont。

--from
from参数指定授权转出的账户地址。

--to
to参数指定转授权入的账户地址。

--amount
amount参数指定授权转账金额。注意：由于ONT的精度是1，因此如果输入的是个浮点值，那么小数部分的值会被丢弃；ONG的精度为9，因此超出9位的小数部分将会被丢弃。

**授权转账**

```
./ontology asset approve --from=<address|index|label> --to=<address|index|label> --amount=XXX --asset=ont
```

### 3.4 查看授权转帐余额

授权用户转账后，用户可以根据需要分多次在授权额度内执行转账操作。查看授权转帐余额命令可以查看到未转账的余额。

#### 3.4.1 查看授权转帐余额参数

--wallet, -w
wallet指定转出账户钱包路径，默认值为:"./wallet.dat"

--asset
asset参数指定转账的资产类型，ont表示ONT，ong表示ONG。默认值为ont。

--from
from参数指定授权转出账户地址。

--to
to参数指定授权转入账户地址。

**查看授权转帐余额**

```
./ontology asset allowance --from=<address|index|label> --to=<address|index|label>
```

### 3.5 从授权账户中转账

通过用户授权后，可以从授权账户中转帐。

#### 3.5.1 从授权账户中转账参数
--wallet, -w
wallet指定执行授权转账账户的钱包路径，默认值为:"./wallet.dat"

--gasprice
gasprice参数指定转账交易的gas price。交易的gas price不能小于接收节点交易池设置的最低gas price，否则交易会被拒绝。默认值为0。当交易池中有交易在排队等待打包进区块时，交易池会按照gas price由高到低排序，gas price高的交易会被优先处理。

--gaslimit
gaslimit参数指定转账交易的gas limit。交易的gas limit不能小于接收节点交易池设置的最低gas limit，否则交易会被拒绝。gasprice * gaslimit 为账户实际支付的ONG 费用。 默认值为30000。

--asset
asset参数指定转账的资产类型，ont表示ONT，ong表示ONG。默认值为ont。

--from
from参数指定授权转出账户地址。

--to
to参数指定转授权入账户地址。

--sender
sender参数指定实际执行授权转账的账户地址。如果没有指定sender参数，sender参数默认使用to参数的指定的账户地址。

--amount
amount参数指定转账金额，转账金额不能大于授权转账余额，否则交易会执行失败。注意：由于ONT的精度是1，因此如果输入的是个浮点值，那么小数部分的值会被丢弃；ONG的精度为9，因此超出9位的小数部分将会被丢弃。

**从授权账户中转账**

```
./ontology asset transferfrom --from=<address|index|label> --to=<address|index|label> --sender=<address|index|label> --amount=XXX
```

### 3.6 查看未解绑的ONG余额

ONG采用定时解绑策略解除绑定在ONT上的ONG。使用如下命令可以查看到当前账户未解绑的ONG余额。

```
./ontology asset unboundong <address|index|label>
```
### 3.7 提取解绑的ONG

使用提取命令可以提取当前所有未解绑的ONG。

#### 3.7.1 提取解绑的ONG参数

--wallet, -w
wallet参数指定提取账户的钱包路径，默认值为:"./wallet.dat"

--gasprice
gasprice参数指定转账交易的gas price。交易的gas price不能小于接收节点交易池设置的最低gas price，否则交易会被拒绝。默认值为0。当交易池中有交易在排队等待打包进区块时，交易池会按照gas price由高到低排序，gas price高的交易会被优先处理。

--gaslimit
gaslimit参数指定转账交易的gas limit。交易的gas limit不能小于接收节点交易池设置的最低gas limit，否则交易会被拒绝。gasprice * gaslimit 为账户实际支付的ONG 费用。 默认值为30000。

**提取解绑的ONG**
```
./ontology asset withdrawong <address|index|label>
```
## 4 查询信息

查询信息命令可以查询区块、交易以及交易执行等信息。使用./ontology info block --help 命令可以查看帮助信息。

### 4.1 查询区块信息

```
./ontology info block <height|blockHash>
```
可以通过区块高度或者区块Hash 查询区块信息。

### 4.2 查询交易信息

```
./ontology info tx <TxHash>
```
可以通过交易Hash查询交易信息。

### 4.3 查询交易执行信息

```
./ontology info status <TxHash>
```
可以通过交易Hash查询交易的执行信息，返回示例如下：

```
{
   "TxHash": "4c00674d96b1d3d2c8152b905cae6f87fff0ec8acf28ca3e7465aac59de814a1",
   "State": 1,
   "GasConsumed": 0,
   "Notify": [
      {
         "ContractAddress": "ff00000000000000000000000000000000000001",
         "States": [
            "transfer",
            "TA587BCw7HFwuUuzY1wg2HXCN7cHBPaXSe",
            "TA5gYXCSiUq9ejGCa54M3yoj9kfMv3ir4j",
            10
         ]
      }
   ]
}
```
其中，State表示交易执行结果，State的值为1，表示交易执行成功，State值为0时，表示执行失败。GasConsumed表示交易执行消耗的ONG。Notify表示交易执行时输出的Event log。不同的交易可能会输出不同的Event log。

## 5、智能合约

智能合约操作支持NeoVM智能合约的部署，以及NeoVM智能合约的预执行和执行。

### 5.1 智能合约部署

智能部署前需要把在NeoVM合约编译器如：[SmartX](http://smartx.ont.io) 上编译好的Code，保存在本地的一个文本文件中。

#### 5.1.1 智能合约部署参数

--wallet, -w
wallet参数指定部署智能合约的账户钱包路径。默认值："./wallet.dat"。

--account, -a
account参数指定部署合约的账户。

--gasprice
gasprice参数指定部署合约交易的gas price。交易的gas price不能小于接收节点交易池设置的最低gas price，否则交易会被拒绝。默认值为0。当交易池中有交易在排队等待打包进区块时，交易池会按照gas price由高到低排序，gas price高的交易会被优先处理。

--gaslimit
gaslimit参数指定部署合约交易的gas limit。交易的gas limit不能小于接收节点交易池设置的最低gas limit，否则交易会被拒绝。gasprice * gaslimit 为账户实际支付的ONG 费用。

**对于合约部署，gaslimit 值必须大于10000000，同时账户中必须保有足够的ONG余额。**

--needstore
needstore参数指定智能合约属否需要使用持久化存储，如果需要使用则需要带上该参数。默认为不使用。

--code
code参数指定智能合约代码路径。

--name
name参数指定智能合约的名称。

--version
version参数指定智能合约的版本号。

--author
author参数指定智能合约的作者信息。

--email
emial参数指定智能合约的联系人电子邮件。

--desc
desc参数可以指定智能合约的描述信息。

--prepare, -p
prepare参数用于预部署合约, 预部署不会把合约部署到Ontology上， 也不会消耗人任何ONG。通过预部署合约，用户可以知道当前合约部署所需要消耗的gas limit。

**智能合约部署**

```
./ontology contract deploy --name=xxx --code=xxx --author=xxx --desc=xxx --email=xxx --needstore --gaslimit=100000000
```

部署后会返回部署交易的TxHash以及合约地址，如：

```
Deploy contract:
  Contract Address:806fbee1fcfb554af47844edd4d4ce2918737747
  TxHash:99d719f51837acfa48f9cd2a21983fb993bc8d5a763b497802f7b872be2338fe
```

可以通过 ./ontology info status <TxHash> 命令查询合约执行状态。如果返回错误如：UNKNOWN TRANSACTION时，表示交易没有落帐，有可能交易还在交易池中排队等待被打包，也有可能表示交易因为gaslimit或者时gasprice设置过低，导致交易被拒绝。

如果返回的执行状态State等于0，表示交易执行失败，如果State等于1，表示交易执行成功，合约被成功部署。如：

```
Transaction states:
{
   "TxHash": "99d719f51837acfa48f9cd2a21983fb993bc8d5a763b497802f7b872be2338fe",
   "State": 1,
   "GasConsumed": 0,
   "Notify": []
}
```

Contract Address为根据合约Code生成的合约地址。

### 5.2 智能合约执行

NeoVM智能合约参数类型支持array、bytearray、string、int以及bool类型。其中array表示对象数组，数组元素可以是NeoVM支持的任意数量、任意类型的值；bytearray表示字节数组，输入时需要将byte数组用十六进制编码成字符串，如 []byte("HelloWorld") 编码成：48656c6c6f576f726c64；string表示字符串字面值；int表示整数，由于NeoVM虚拟机不支持浮点数值，因此需要将浮点数转换成整数；bool表示布尔型变量，用true，false表示。

在Ontology cli中，使用前缀法构造输入参数，参数前使用类型标识标注类型，如字符串参数表示为 string:hello; 整数参数表示为 int:10; 布尔类型参数表示为 bool:true等。多个参数使用","分隔。对象数组array类型用"[ ]"表示数组元素范围，如 [int:10,string:hello,bool:true]。

输入参数示例：

```
string:methodName,[string:arg1,int:arg2]
```

#### 5.2.1 智能合约执行参数

--wallet, -w
wallet参数指定智能合约执行的账户钱包路径。默认值："./wallet.dat"。

--account, -a
account参数指定执行合约的账户。

--gasprice
gasprice参数指定部署合约交易的gas price。交易的gas price不能小于接收节点交易池设置的最低gas price，否则交易会被拒绝。默认值为0。当交易池中有交易在排队等待打包进区块时，交易池会按照gas price由高到低排序，gas price高的交易会被优先处理。

--gaslimit
gaslimit参数指定部署合约交易的gas limit。交易的gas limit不能小于接收节点交易池设置的最低gas limit，否则交易会被拒绝。gasprice * gaslimit 为账户实际支付的ONG 费用。

--address
address参数指定调用的合约地址

--params
params参数用于输入合约调用的参数，需要按照上面的说明编码输入参数。

--prepare, -p
prepare参数表示当前为预执行，执行交易不会被打包到区块中，也不会消耗任何ONG。预执行会返回合约方法的返回值，同时还会试算当前调用需要的gas limit。

--return
return参数用于配合--prepare参数使用，在预执行时通过--return参数标注的返回值类型来解析合约返回返回值，否则输出合约方法调用时返回的原始值。多个返回值类型用","分隔，如 string,int

**智能合约预执行**

```
./ontology contract invoke --address=XXX --params=XXX --return=XXX --p
```
返回示例：

```
Contract invoke successfully
Gas consumed:30000
Return:0
```
**智能合约执行**

```
./ontology contract invoke --address=XXX --params=XXX --gaslimit=XXX
```

智能合约在执行之前，可以通过预执行，试算出当前执行所需要的gas limit，以避免ONG余额不足导致执行失败。

### 5.3 直接执行智能合约字节码

智能合约部署后，cli支持直接执行NeoVM Code。

#### 5.3.1 直接执行智能合约字节码参数

--wallet, -w
wallet参数指定智能合约执行的账户钱包路径。默认值："./wallet.dat"。

--account, -a
account参数指定执行合约的账户。

--gasprice
gasprice参数指定部署合约交易的gas price。交易的gas price不能小于接收节点交易池设置的最低gas price，否则交易会被拒绝。默认值为0。当交易池中有交易在排队等待打包进区块时，交易池会按照gas price由高到低排序，gas price高的交易会被优先处理。

--gaslimit
gaslimit参数指定部署合约交易的gas limit。交易的gas limit不能小于接收节点交易池设置的最低gas limit，否则交易会被拒绝。gasprice * gaslimit 为账户实际支付的ONG 费用。

--prepare, -p
prepare参数表示当前为预执行，执行交易不会被打包到区块中，也不会消耗任何ONG。预执行会返回合约方法的返回值，同时还会试算当前调用需要的gas limit。

--code
code参数指定可执行的智能合约代码路径。

#### 5.3.2 直接执行智能合约字节码

```
./ontology contract invokeCode --code=XXX --gaslimit=XXX
```

## 6、区块导入导出

Ontology Cli支持导出本地节点的区块数据到一个压缩文件中，生成的压缩文件可以再导入其它Ontology节点中。出于安全考虑，导入的区块数据文件请确保是从可信的来源获取的。

### 6.1 导出区块

#### 6.1.1 导出区块参数

--file
file参数指定导出的文件路径。默认值为：blocks.dat

--height
height参数指定导出的终止区块高度。当本地节点的当前区块高度大于导出的终止高度，大于的部分不会被导出。height等于0，表示导出当前节点的所有区块。默认值为0。

--speed
speed参数指定导出速度。分别用h表示high，m表示middle，l表示low。默认值为m。

区块导出

```
./ontology export
```

### 6.2 导入区块

#### 6.2.1 导入区块参数

--importheight
importheight 参数指定导入的目标区块高度。如果importheight指定的区块高度小于区块文件的最大高度时，只导入到importheight指定的高度，剩余的区块会停止导入。默认值为0，表示导入所有的区块。

--importfile
importfile 参数配合--import使用，用于区块导入时指定导入文件的路径。默认值为"./blocks.dat"。

导入区块

```
./ontology import
```
