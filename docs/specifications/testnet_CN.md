
<h1 align="center">Ontology </h1>
<p align="center" class="version">Version 0.7.0 </p>

[![GoDoc](https://godoc.org/github.com/ontio/ontology?status.svg)](https://godoc.org/github.com/ontio/ontology)
[![Go Report Card](https://goreportcard.com/badge/github.com/ontio/ontology)](https://goreportcard.com/report/github.com/ontio/ontology)
[![Travis](https://travis-ci.org/ontio/ontology.svg?branch=master)](https://travis-ci.org/ontio/ontology)
[![Gitter](https://badges.gitter.im/Join%20Chat.svg)](https://gitter.im/ontio/ontology?utm_source=badge&utm_medium=badge&utm_campaign=pr-badge)

[English](testnet.md) | 中文

## 服务器部署
### 选择网络
ontology的运行支持以下3钟方式

* 连接到公开测试网(Polaris)
* 单机部署
* 多机部署

#### 连接到公开测试网(Polaris)
1.创建钱包
- 通过命令行程序，分别创建节点运行所需的钱包文件wallet.dat 
    ```
    $ ./ontology account add -d
    use default value for all options
    Enter a password for encrypting the private key:
    Re-enter password:
    
    Create account successfully.
    Address:  TA9TVuR4Ynn4VotfpExY5SaEy8a99obFPr
    Public key: 120202a1cfbe3a0a04183d6c25ceff1e34957ace6e4899e4361c2e1a2bc3c817f90936
    Signature scheme: SHA256withECDSA
    ```
    配置的例子如下：
    - 目录结构
    ```shell
    $ tree
    └── ontology
        ├── ontology
        └── wallet.dat
    ```     

#### 单机部署配置

在单机上创建一个目录，在目录下存放以下文件：
- 默认配置文件`config.json`
- 节点程序 + 节点控制程序 `ontology`
- 钱包文件`wallet.dat`
把source根目录下的config-solo.config配置文件的内容复制到config.json内，修改Bookkeepers配置，替换钱包公钥，然后启动节点即可。钱包公钥可以使用`$ ./ontology account list -v` 获取

单机配置的例子如下：
- 目录结构
    ```shell
    $ tree
    └── node
        ├── config.json
        ├── ontology
        └── wallet.dat
    ```
- config.json中Bookkeepers的配置：
    ```
    "Bookkeepers": [ "1202021c6750d2c5d99813997438cee0740b04a73e42664c444e778e001196eed96c9d" ],
    ```

#### 多机部署配置

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
        ```
        $ ./ontology account add -d
        use default value for all options
        Enter a password for encrypting the private key:
        Re-enter password:
        
        Create account successfully.
        Address:  TA9TVuR4Ynn4VotfpExY5SaEy8a99obFPr
        Public key: 120202a1cfbe3a0a04183d6c25ceff1e34957ace6e4899e4361c2e1a2bc3c817f90936
        Signature scheme: SHA256withECDSA
        ```

5. 记账人配置
    - 为每个节点创建钱包时会显示钱包的公钥信息，将所有节点的公钥信息分别填写到每个节点的配置文件的`Bookkeepers`项中
    
        注：每个节点的钱包公钥信息也可以通过命令行程序查看：
    
        ```
        $ ./ontology account list -v
        * 1     TA9TVuR4Ynn4VotfpExY5SaEy8a99obFPr
                Signature algorithm: ECDSA
                Curve: P-256
                Key length: 256 bit
                Public key: 120202a1cfbe3a0a04183d6c25ceff1e34957ace6e4899e4361c2e1a2bc3c817f90936 bit
                Signature scheme: SHA256withECDSA
        ```


多机部署配置完成，每个节点目录结构如下

```shell
$ ls
config.json ontology wallet.dat
```

一个配置文件片段可以参考根目录下的config-dbft.json文件。

### 运行
以任意顺序运行每个节点node程序，并在出现`Password:`提示后输入节点的钱包密码

```shell
$ ./ontology
$ - 输入你的钱包口令
```

了解更多请运行 `./ontology --help`.

### ONT转账调用示例
  contract:合约地址； - from: 转出地址； - to: 转入地址； - value: 资产转移数量；
```shell
  ./ontology asset transfer --caddr=ff00000000000000000000000000000000000001 --value=500 --from  TA6nAAdX77wcsAnuBQxG61zXg3vJUAPpgk  --to TA6Hsjww86b9KBbXFyKEayMcVVafoTGH4K  --password=xxx
```
如果成功调用会返回如下event:
```
[
	{
		"ContractAddress": "ff00000000000000000000000000000000000001",
		"TxHash": "e0ba3d5807289eac243faceb1a2ac63e8dee4eba208ceac193b0bd606861b729",
		"States": [
			"transfer",
			"TA6nAAdX77wcsAnuBQxG61zXg3vJUAPpgk",
			"TA6Hsjww86b9KBbXFyKEayMcVVafoTGH4K",
			500
		]
	}
]