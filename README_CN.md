
<h1 align="center">Ontology </h1>
<h4 align="center">Version 0.7.0 </h4>

[![GoDoc](https://godoc.org/github.com/ontio/ontology?status.svg)](https://godoc.org/github.com/ontio/ontology)
[![Go Report Card](https://goreportcard.com/badge/github.com/ontio/ontology)](https://goreportcard.com/report/github.com/ontio/ontology)
[![Travis](https://travis-ci.org/ontio/ontology.svg?branch=master)](https://travis-ci.org/ontio/ontology)
[![Gitter](https://badges.gitter.im/Join%20Chat.svg)](https://gitter.im/ontio/ontology?utm_source=badge&utm_medium=badge&utm_campaign=pr-badge)

[English](README.md) | 中文

欢迎来到Ontology的源码库！ 

Ontology致力于创建一个组件化、可自由配置、跨链支持、高性能、横向可扩展的区块链底层基础设施。 让部署及调用去中心化应用变得更加非常简单。

目前代码还处于内部测试阶段，但处于快速的开发过程中，master代码可能是不稳定的，稳定的版本可以在releases中下载。

公开的测试网可以在下面找到，也非常欢迎及希望能有更多的开发者加入到Ontology中来。

## 特性

* 可扩展的轻量级通用智能合约
* 可扩展的WASM合约的支持
* 跨链交互协议（进行中）
* 多种加密算法支持 
* 高度优化的交易处理速度
* P2P连接链路加密(可选择模块)
* 多种共识算法支持 (VBFT/DBFT/SBFT/PoW/SOLO...)
* 快速的区块生成时间

## 目录

* [构建开发环境](#构建开发环境)
* [获取ontology](#获取ontology)
    * [从源码获取](#从源码获取)
    * [从release获取](#从release获取)
* [服务器部署](#服务器部署)
    * [选择网络](#选择网络)
        * [公开测试网Polaris同步节点部署](#公开测试网polaris同步节点部署)
        * [单机部署配置](#单机部署配置)
        * [多机部署配置](#多机部署配置)
    * [运行](#运行)
    * [ONT转账调用示例](#ont转账调用示例)
* [贡献代码](#贡献代码)
* [开源社区](#开源社区)
    * [网站](#网站)
    * [Discord开发者社区](#discord开发者社区)
* [许可证](#许可证)

## 构建开发环境
成功编译ontology需要以下准备：

* Golang版本在1.9及以上
* 安装第三方包管理工具glide
* 正确的Go语言开发环境
* Golang所支持的操作系统

## 获取ontology
### 从源码获取
克隆ontology仓库到 **$GOPATH/src/github.com/ontio** 目录

```shell
$ git clone https://github.com/ontio/ontology.git
```
或者
```shell
$ go get github.com/ontio/ontology
```

用第三方包管理工具glide拉取依赖库

````shell
$ cd $GOPATH/src/github.com/ontio/ontology
$ glide install
````

用make编译源码

```shell
$ make
```

成功编译后会生成两个可以执行程序

* `ontology`: 节点程序/以命令行方式提供的节点控制程序

### 从release获取
You can download at [release page](https://github.com/ontio/ontology/releases).

## 服务器部署
### 选择网络
ontology的运行支持以下3种方式

* 公开测试网Polaris同步节点部署
* 单机部署
* 多机部署

#### 公开测试网Polaris同步节点部署
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

2.启动./ontology节点
  * 不需要config.json文件，会使用默认配置启动节点

**注意**：钱包文件的格式有变化，旧文件无法继续使用，请重新生成新的钱包文件。

#### 单机部署配置

在单机上创建一个目录，在目录下存放以下文件：
- 节点程序 + 节点控制程序 `ontology`
- 钱包文件`wallet.dat`

使用命令 `$ ./ontology --testmode` 即可启动单机版的测试网络。

单机配置的例子如下：
- 目录结构

    ```shell
    $ tree
    └── node
        ├── ontology
        └── wallet.dat
    ```

#### 多机部署配置

网络环境下，最少需要4个节点（共识节点）完成部署。
我们可以通过修改默认的配置文件`config.json`进行快速部署。

1. 将相关文件复制到目标主机，包括：
    - 默认配置文件`config.json`
    - 节点程序`ontology`

2. 种子节点配置
    - 在4个主机中选出至少一个做种子节点，并将种子节点地址分别填写到每个配置文件的`SeelList`中，格式为`种子节点IP地址 + 种子节点NodePort`

3. 创建钱包文件
    - 通过命令行程序，在每个主机上分别创建节点运行所需的钱包文件wallet.dat

    ```shell
    $ ./ontology account add -d
    use default value for all options
    Enter a password for encrypting the private key:
    Re-enter password:

    Create account successfully.
    Address:  TA9TVuR4Ynn4VotfpExY5SaEy8a99obFPr
    Public key: 120202a1cfbe3a0a04183d6c25ceff1e34957ace6e4899e4361c2e1a2bc3c817f90936
    Signature scheme: SHA256withECDSA
   ```

4. 记账人配置
    - 为每个节点创建钱包时会显示钱包的公钥信息，将所有节点的公钥信息分别填写到每个节点的配置文件的`Bookkeepers`项中

    注：每个节点的钱包公钥信息也可以通过命令行程序查看：

    ```shell
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
$ ./ontology --nodeport=20338 --rpcport=20336
$ - 输入你的钱包口令
```

了解更多请运行 `./ontology --help`.

### ONT转账调用示例
   - from: 转出地址； - to: 转入地址； - amount: 资产转移数量；
   from参数可以不指定，如果不指定则使用默认账户。

```shell
  ./ontology asset transfer  --to=TA4Xe9j8VbU4m3T1zEa1uRiMTauiAT88op --amount=10
```

执行完后会输出：

```
Transfer ONT
From:TA6edvwgNy3c1nBHgmFj8KrgQ1JCJNhM3o
To:TA4Xe9j8VbU4m3T1zEa1uRiMTauiAT88op
Amount:10
TxHash:10dede8b57ce0b272b4d51ab282aaf0988a4005e980d25bd49685005cc76ba7f
```
其中TxHash是转账交易的交易HASH，可以通过这个HASH查询转账交易的直接结果。
出于区块链出块时间的限制，提交的转账请求不会马上执行，需要等待至少一个区块时间，等待记账节点打包交易。

### 查询转账结果示例

--hash:指定查询的转账交易hash
```shell
./ontology asset status --hash=10dede8b57ce0b272b4d51ab282aaf0988a4005e980d25bd49685005cc76ba7f
```
查询结果：
```shell
Transaction:transfer success
From:TA6edvwgNy3c1nBHgmFj8KrgQ1JCJNhM3o
To:TA4Xe9j8VbU4m3T1zEa1uRiMTauiAT88op
Amount:10
```

### 查询账户余额示例

--address:账户地址

```shell
./ontology asset balance --address=TA4Xe9j8VbU4m3T1zEa1uRiMTauiAT88op
```
查询结果：
```shell
BalanceOf:TA4Xe9j8VbU4m3T1zEa1uRiMTauiAT88op
ONT:10
ONG:0
ONGApprove:0
```

## 贡献代码

请您以签过名的commit发送pull request请求，我们期待您的加入！
您也可以通过邮件的方式发送你的代码到开发者邮件列表，欢迎加入Ontology邮件列表和开发者论坛。

另外，在您想为本项目贡献代码时请提供详细的提交信息，格式参考如下：

  Header line: explain the commit in one line (use the imperative)

  Body of commit message is a few lines of text, explaining things
  in more detail, possibly giving some background about the issue
  being fixed, etc etc.

  The body of the commit message can be several paragraphs, and
  please do proper word-wrap and keep columns shorter than about
  74 characters or so. That way "git log" will show things
  nicely even when it's indented.

  Make sure you explain your solution and why you're doing what you're
  doing, as opposed to describing what you're doing. Reviewers and your
  future self can read the patch, but might not understand why a
  particular solution was implemented.

  Reported-by: whoever-reported-it
  Signed-off-by: Your Name <youremail@yourhost.com>

## 开源社区

### 网站

- https://ont.io/

### Discord开发者社区

- https://discord.gg/4TQujHj/

## 许可证

Ontology遵守GNU Lesser General Public License, 版本3.0。 详细信息请查看项目根目录下的LICENSE文件。
