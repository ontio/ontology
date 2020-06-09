
<h1 align="center">Ontology </h1>
<h4 align="center">Version 1.8.0 </h4>

[![GoDoc](https://godoc.org/github.com/ontio/ontology?status.svg)](https://godoc.org/github.com/ontio/ontology)
[![Go Report Card](https://goreportcard.com/badge/github.com/ontio/ontology)](https://goreportcard.com/report/github.com/ontio/ontology)
[![Travis](https://travis-ci.com/ontio/ontology.svg?branch=master)](https://travis-ci.com/ontio/ontology)
[![Discord](https://img.shields.io/discord/102860784329052160.svg)](https://discord.gg/gDkuCAq)

[English](README.md) | 中文

欢迎来到Ontology的源码库！

Ontology致力于创建一个组件化、可自由配置、跨链支持、高性能、横向可扩展的区块链底层基础设施。 让部署及调用去中心化应用变得更加非常简单。

Ontology MainNet 已经在2018年6月30日成功上线。<br>
但很多新的功能还处于快速的开发过程中，master分支的代码可能是不稳定的，稳定的版本可以在[releases](https://github.com/ontio/ontology/releases)中下载。

公开的主网和测试网都可以在下面找到，也非常欢迎及希望能有更多的开发者加入到Ontology中来。

* [特性](#特性)
* [构建开发环境](#构建开发环境)
* [获取ontology](#获取ontology)
    * [从release获取](#从release获取)
    * [从源码获取](#从源码获取)
* [运行ontology](#运行ontology)
    * [主网同步节点](#主网同步节点)
    * [公开测试网Polaris同步节点](#公开测试网polaris同步节点)
    * [测试模式](#测试模式)
    * [使用docker运行](#使用docker运行)
* [使用示例](#使用示例)
    * [ONT转账调用示例](#ont转账调用示例)
	* [查询转账结果示例](#查询转账结果示例)
	* [查询账户余额示例](#查询账户余额示例)
* [贡献代码](#贡献代码)
* [开源社区](#开源社区)
    * [网站](#网站)
    * [Discord开发者社区](#discord开发者社区)
* [许可证](#许可证)

## 特性

* 可扩展的轻量级通用智能合约
* 可扩展的WASM合约的支持
* 跨链交互协议（进行中）
* 多种加密算法支持
* 高度优化的交易处理速度
* P2P连接链路加密(可选择模块)
* 多种共识算法支持 (VBFT/DBFT/SBFT/PoW/SOLO...)
* 快速的区块生成时间

## 构建开发环境
成功编译ontology需要以下准备：

* Golang版本在1.12及以上
* 正确的Go语言开发环境
* Golang所支持的操作系统

## 获取ontology

### 从release获取
- 你可以通过命令 ` curl https://dev.ont.io/ontology_install | sh ` 获取最新的ontology版本
- 你也可以从[下载页面](https://github.com/ontio/ontology/releases)获取.

### 从源码获取

```shell
$ git clone https://github.com/ontio/ontology.git
$ cd ontology
$ make all
```

成功编译后会生成两个可以执行程序

* `ontology`: 节点程序/以命令行方式提供的节点控制程序
* `tools/sigsvr`: (可选)签名服务 - sigsvr是一个签名服务的server以满足一些特殊的需求。详细的文档可以在[这里](./docs/specifications/sigsvr_CN.md)参考

## 运行ontology

### 主网同步节点

直接启动Ontology

```
./ontology
```
然后你可以连接上主网了。

### 公开测试网Polaris同步节点

直接启动Ontology

```
./ontology --networkid 2
```
然后你可以连接上公共测试网了。

### 测试模式

在单机上创建一个目录，在目录下存放以下文件：
- 节点程序`ontology`
- 钱包文件`wallet.dat` （注：`wallet.dat`可通过`./ontology account add -d`生成）

使用命令 `$ ./ontology --testmode` 即可启动单机版的测试网络。

单机配置的例子如下：
- 目录结构

    ```shell
    $ tree
    └── node
        ├── ontology
        └── wallet.dat
    ```

### 使用docker运行

请确保机器上已安装有docker环境。

1. 编译docker镜像

    - 在下载好的源码根目录下，运行`make docker`命令，这将编译好ontology的docker镜像

2. 运行ontology镜像

    - 使用命令`docker run ontio/ontology`运行ontology；

    - 如果需要使镜像运行时接受交互式键盘输入，则使用`docker run -ti ontio/ontology`命令启动镜像即可；

    - 如果需要保留镜像每次运行时的数据，可以参考docker的数据持久化功能（例如 valume）；

    - 如果需要使用ontology参数，则在`docker run ontio/ontology`后面直接加参数即可，例如`docker run ontio/ontology --networkid 2`，具体的ontology命令
    行参数可以参考[这里](./docs/specifications/cli_user_guide_CN.md)。

## 使用示例

### ONT转账调用示例
   - from: 转出地址； - to: 转入地址； - amount: 资产转移数量；
      from参数可以不指定，如果不指定则使用默认账户。

```shell
  ./ontology asset transfer  --from=ARVVxBPGySL56CvSSWfjRVVyZYpNZ7zp48 --to=AaCe8nVkMRABnp5YgEjYZ9E5KYCxks2uce --amount=10
```

执行完后会输出：

```shell
Transfer ONT
  From:ARVVxBPGySL56CvSSWfjRVVyZYpNZ7zp48
  To:AaCe8nVkMRABnp5YgEjYZ9E5KYCxks2uce
  Amount:10
  TxHash:437bff5dee9a1894ad421d55b8c70a2b7f34c574de0225046531e32faa1f94ce
```
其中TxHash是转账交易的交易HASH，可以通过这个HASH查询转账交易的直接结果。
出于区块链出块时间的限制，提交的转账请求不会马上执行，需要等待至少一个区块时间，等待记账节点打包交易。

如果需要转ONG，可以使用参数 -- asset = ong。注意，ONT最少单位是1，而ONG则有9位小数点。

```shell
./ontology asset transfer --from=ARVVxBPGySL56CvSSWfjRVVyZYpNZ7zp48 --to=ARVVxBPGySL56CvSSWfjRVVyZYpNZ7zp48 --amount=95.479777254 --asset=ong
```
执行完后会输出：

```shell
Transfer ONG
  From:ARVVxBPGySL56CvSSWfjRVVyZYpNZ7zp48
  To:AaCe8nVkMRABnp5YgEjYZ9E5KYCxks2uce
  Amount:95.479777254
  TxHash:e4245d83607e6644c360b6007045017b5c5d89d9f0f5a9c3b37801018f789cc3
```

注意，Ontology cli中，所有用到账户的地址的地方，都支持账户索引和账户标签。账户索引是账户在钱包中的序号，从1开始。标签是可以在创建账户的时候指定一个唯一的别名。如：

```shell
./ontology asset transfer --from=1 --to=2 --amount=10
```

### 查询转账结果示例

```shell
./ontology info status <TxHash>
```

如：

```shell
./ontology info status e4245d83607e6644c360b6007045017b5c5d89d9f0f5a9c3b37801018f789cc3
```

查询结果：
```shell
Transaction states:
{
   "TxHash": "e4245d83607e6644c360b6007045017b5c5d89d9f0f5a9c3b37801018f789cc3",
   "State": 1,
   "GasConsumed": 0,
   "Notify": [
      {
         "ContractAddress": "0200000000000000000000000000000000000000",
         "States": [
            "transfer",
            "ARVVxBPGySL56CvSSWfjRVVyZYpNZ7zp48",
            "AaCe8nVkMRABnp5YgEjYZ9E5KYCxks2uce",
            95479777254
         ]
      }
   ]
}
```

### 查询账户余额示例

```shell
./ontology asset balance <address|index|label>
```
如：

```shell
./ontology asset balance ARVVxBPGySL56CvSSWfjRVVyZYpNZ7zp48
```
或者

```shell
./ontology asset balance 1
```
查询结果：
```shell
BalanceOf:ARVVxBPGySL56CvSSWfjRVVyZYpNZ7zp48
  ONT:989979697
  ONG:28165900
```

进一步的示例可以参考[文档中心](https://ontio.github.io/documentation/)

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
