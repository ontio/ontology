
<h1 align="center">Ontology </h1>
<h4 align="center">Version 0.7.0 </h4>

[![GoDoc](https://godoc.org/github.com/ontio/ontology?status.svg)](https://godoc.org/github.com/ontio/ontology)
[![Go Report Card](https://goreportcard.com/badge/github.com/ontio/ontology)](https://goreportcard.com/report/github.com/ontio/ontology)
[![Travis](https://travis-ci.org/ontio/ontology.svg?branch=master)](https://travis-ci.org/ontio/ontology)
[![Gitter](https://badges.gitter.im/Join%20Chat.svg)](https://gitter.im/ontio/ontology?utm_source=badge&utm_medium=badge&utm_campaign=pr-badge)

[English](testnet.md) | 中文

# 服务器部署
成功运行ontology可以通过以下两种方式进行部署

* 单机部署
* 多机部署
 * 在公共测试网上部署节点

### 单机部署配置

在单机上创建一个目录，在目录下存放以下文件：
- 默认配置文件`config.json`
- 节点程序 + 节点控制程序 `ontology`
- 钱包文件`wallet.dat`
把source根目录下的config-solo.config配置文件的内容复制到config.json内，修改Bookkeepers配置，替换钱包公钥，然后启动节点即可。钱包公钥可以使用`$ ./ontology wallet show --name=wallet.dat` 获取

单机配置的例子如下：
- 目录结构

```shell
$ tree
└── node
    ├── config.json
    ├── ontology
    └── wallet.dat
```
config.json中Bookkeepers的配置：
```
"Bookkeepers": [ "1202021c6750d2c5d99813997438cee0740b04a73e42664c444e778e001196eed96c9d" ],
```

### 多机部署配置

网络环境下，最少需要4个节点（共识节点）完成部署。
我们可以通过修改默认的配置文件`config.json`进行快速部署。

1. 将相关文件复制到目标主机，包括：
    - 默认配置文件`config.json`
    - 节点程序`ontology`

2. 设置每个节点网络连接的端口号（推荐不做修改，使用默认端口配置）
    - `NodePort`为的P2P连接端口号（默认20338）
    - `HttpJsonPort`和`HttpLocalPort`为RPC端口号（默认为20336，20337）

3. 种子节点配置
    - 在4个主机中选出至少一个做种子节点，并将种子节点地址分别填写到每个配置文件的`SeelList`中，格式为`种子节点IP地址 + 种子节点NodePort`

4. 创建钱包文件
    - 通过命令行程序，在每个主机上分别创建节点运行所需的钱包文件wallet.dat 
      
        `$ ./ontology wallet -create --name="wallet.dat"` 

        注：通过-p参数设置钱包密码

5. 记账人配置
    - 为每个节点创建钱包时会显示钱包的公钥信息，将所有节点的公钥信息分别填写到每个节点的配置文件的`Bookkeepers`项中
    
        注：每个节点的钱包公钥信息也可以通过命令行程序查看：
    
        `$ ./ontology wallet show --name=wallet.dat` 


多机部署配置完成，每个节点目录结构如下

```shell
$ ls
config.json ontology wallet.dat
```

一个配置文件片段如下, 可以参考根目录下的config.json文件。

### 在公共测试网上部署节点(default config)
按照以下配置文件启动可以连接到ont目前的测试网络。

```shell
$ cat config.json
{
  "Configuration": {
    "Magic": 7630401,
    "Version": 23,
    "SeedList": [
     "139.219.108.204:20338",
     "139.219.111.50:20338",
     "139.219.69.70:20338",
     "40.125.165.118:20338"
    ],
    "Bookkeepers": [
"1202021c6750d2c5d99813997438cee0740b04a73e42664c444e778e001196eed96c9d",
"12020339541a43af2206358714cf6bd385fc9ac8b5df554fec5497d9e947d583f985fc",
"120203bdf0d966f98ff4af5c563c4a3e2fe499d98542115e1ffd75fbca44b12c56a591",
"1202021401156f187ec23ce631a489c3fa17f292171009c6c3162ef642406d3d09c74d"
    ],
    "HttpRestPort": 20334,
    "HttpWsPort":20335,
    "HttpJsonPort": 20336,
    "HttpLocalPort": 20337,
    "NodePort": 20338,
    "NodeConsensusPort": 20339,
    "PrintLevel": 1,
    "IsTLS": false,
    "MaxTransactionInBlock": 60000,
    "MultiCoreNum": 4
  }
}

```

### 运行
以任意顺序运行每个节点node程序，并在出现`Password:`提示后输入节点的钱包密码

```shell
$ ./ontology
$ - 输入你的钱包口令
```

了解更多请运行 `./ontology --help`.

# 简单示例
## 合约
[合约Guide](https://github.com/ontio/documentation/tree/master/smart-contract-tutorial)

## ONT转账调用示例
  contract:合约地址； - from: 转出地址； - to: 转入地址； - value: 资产转移数量；
```shell
  .\ontology asset transfer --caddr=ff00000000000000000000000000000000000001 --value=500 --from  TA6nAAdX77wcsAnuBQxG61zXg3vJUAPpgk  --to TA6Hsjww86b9KBbXFyKEayMcVVafoTGH4K  --password=xxx
```