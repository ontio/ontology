
<h1 align="center">Ontology </h1>
<p align="center" class="version">Version 1.0.0 </p>

[![GoDoc](https://godoc.org/github.com/ontio/ontology?status.svg)](https://godoc.org/github.com/ontio/ontology)
[![Go Report Card](https://goreportcard.com/badge/github.com/ontio/ontology)](https://goreportcard.com/report/github.com/ontio/ontology)
[![Travis](https://travis-ci.org/ontio/ontology.svg?branch=master)](https://travis-ci.org/ontio/ontology)
[![Gitter](https://badges.gitter.im/Join%20Chat.svg)](https://gitter.im/ontio/ontology?utm_source=badge&utm_medium=badge&utm_campaign=pr-badge)

[English](testnet.md) | 中文

## 服务器部署
### 选择网络
ontology的运行支持以下3钟方式

* 主网同步节点部署
* 连接到公开测试网(Polaris)
* 单机部署
* 多机部署

#### 主网同步节点部署

直接启动Ontology，不需要钱包文件，也不需要配置文件

   ```
	./ontology --networkid 1
   ```

#### 公开测试网Polaris同步节点部署

直接启动Ontology，不需要钱包文件，也不需要配置文件

   ```
	./ontology --networkid 2
   ```

#### 单机部署配置

1.在单机上创建一个目录，在目录下存放以下文件：
- 节点程序 + 节点控制程序 `ontology`
- 钱包文件`wallet.dat`

2.单机配置的例子如下：
- 目录结构
    ```shell
    $ tree
    └── node
        ├── ontology
        └── wallet.dat
    ```
3.启动ontology

   ```
	./ontology  --testmode
   ```
#### 多机部署配置

##### VBFT部署方法

多机环境下，根据VBFT共识算法的要求，最少需要7个节点（共识节点）完成部署。

我们可以通过修改默认的配置文件`config.json`进行快速部署。

1. 生成七个钱包文件，每个钱包文件包含一个账户，共七个账户，分别作为每个节点的记账人。生成账户和钱包的命令为：
	```
	./ontology account add -d -w wallet.dat
	Use default setting '-t ecdsa -b 256 -s SHA256withECDSA' 
		signature algorithm: ecdsa 
		curve: P-256 
		signature scheme: SHA256withECDSA 
	Password:
	Re-enter Password:

	Index: 1
	Label: 
	Address: AXkDGfr9thEqWmCKpTtQYaazJRwQzH48eC
	Public key: 03d7d8c0c4ca2d2bc88209db018dc0c6db28380d8674aff86011b2a6ca32b512f9
	Signature scheme: SHA256withECDSA

	Create account successfully.
	```
	使用-w参数指定生成的钱包文件名

2. 修改config.json，将上一步生成的七个账户的公钥、address分别填入`config.json`中的peers配置的1-7项

3. 将相关文件复制到目标主机，包括：
    - 默认配置文件`config.json`
    - 节点程序`ontology`
    - 钱包文件

4. 设置每个节点网络连接的端口号（推荐不做修改，使用默认端口配置）
    - `NodePort`为的P2P连接端口号（默认20338）
    - `HttpJsonPort`和`HttpLocalPort`为RPC端口号（默认为20336，20337）

5. 种子节点配置
    - 在7个主机中选出至少一个做种子节点，并将种子节点地址分别填写到每个配置文件的`SeelList`中，格式为`种子节点IP地址 + 种子节点NodePort`

##### DBFT部署方法

多机环境下，最少需要4个节点（共识节点）完成部署。
我们可以通过修改默认的配置文件`config-dbft.json`进行快速部署。

1. 将相关文件复制到目标主机，包括：
    - 默认配置文件`config-dbft.json`
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
        Use default setting '-t ecdsa -b 256 -s SHA256withECDSA' 
		signature algorithm: ecdsa 
		curve: P-256 
		signature scheme: SHA256withECDSA 
		Password:
		Re-enter Password:

		Index: 1
		Label: 
		Address: AXkDGfr9thEqWmCKpTtQYaazJRwQzH48eC
		Public key: 03d7d8c0c4ca2d2bc88209db018dc0c6db28380d8674aff86011b2a6ca32b512f9
		Signature scheme: SHA256withECDSA

		Create account successfully.
        ```

5. 记账人配置
    - 为每个节点创建钱包时会显示钱包的公钥信息，将所有节点的公钥信息分别填写到每个节点的配置文件的`Bookkeepers`项中
    
        注：每个节点的钱包公钥信息也可以通过命令行程序查看：
    
        ```
        1	AYiToLDT2yZuNs3PZieXcdTpyC5VWQmfaN (default)
        	Label: 
        	Signature algorithm: ECDSA
        	Curve: P-256
        	Key length: 384 bits
        	Public key: 030e5d50bf585ff5c73464114244b93f04b231862d6bbdfd846be890093b2c1c17
        	Signature scheme: SHA256withECDSA
        ```
6. 修改配置文件名
	- 修改每个节点的`config-dbft.json`为`config.json`
	
#### 多机部署配置完成，每个节点目录结构如下

   ```shell
	$ ls
	config.json ontology wallet.dat
   ```
### 运行
以任意顺序运行每个节点node程序（如果是VBFT，则需要在启动时使用--enableConsensus）， 并在出现`Password:`提示后输入节点的钱包密码

	```shell
	$ ./ontology
	$ - 输入你的钱包口令
	```

了解更多请运行 `./ontology --help`.

### ONT转账调用示例
  - asset: 资产类型["ont"|"ong"] - from: 转出地址； - to: 转入地址； - amount: 资产转移数量；
```shell
  ./ontology asset transfer --amount=500 --from  AYiToLDT2yZuNs3PZieXcdTpyC5VWQmfaN  --to AeoBhZtS8AmGp3Zt4LxvCqhdU4eSGiK44M
```
如果成功调用会返回如下event:
```
Transfer ONT
  From:AYiToLDT2yZuNs3PZieXcdTpyC5VWQmfaN
  To:AeoBhZtS8AmGp3Zt4LxvCqhdU4eSGiK44M
  Amount:1
  TxHash:d232bf1c876fc6fd2f8be90006e523897b1e2082e6df8f08466c56985d24a2a7

Tip:
  Using './ontology info status d232bf1c876fc6fd2f8be90006e523897b1e2082e6df8f08466c56985d24a2a7' to query transaction status
```