
<h1 align="center">Ontology </h1>
<h4 align="center">Version V0.6.0 </h4>

欢迎来到Ontology的源码库！ 

Ontology致力于创建一个组件化、可自由配置、跨链支持、高性能、横向可扩展的区块链底层基础设施。 让部署及调用去中心化应用变得更加非常简单。

目前代码还处于内部测试阶段，但处于快速的开发过程中，master代码可能是不稳定的，稳定的版本可以在releases中下载。

公开的测试网可以在下面找到，也非常欢迎及希望能有更多的开发者加入到ontology中来。

## 特性

* 可扩展的轻量级通用智能合约
* 可扩展的WASM合约的支持
* 跨链交互协议（进行中）
* 多种加密算法支持 
* 高度优化的交易处理速度
* P2P连接链路加密(可选择模块)
* 多种共识算法支持 (VBFT/DBFT/SBFT/SOLO...)
* 快速的区块生成时间

## 目录

* [构建开发环境](#构建开发环境)
* [部署及测试](#部署及测试)
  * [获取ontology](#获取ontology)
    * [从源码获取](#从源码获取)
    * [从release获取](#从release获取)
  * [创建ONT钱包文件](#创建ont钱包文件)
  * [服务器部署](#服务器部署)
    * [单机部署配置](#单机部署配置)
    * [多机部署配置](#多机部署配置)
    * [在公共测试网上部署节点](#在公共测试网上部署节点)
    * [运行](#运行)
* [简单示例](#简单示例)
  * [ONT转账调用示例](#ont转账调用示例)
* [贡献代码](#贡献代码)
* [开源社区](#开源社区)
  * [网站](#网站)
* [许可证](#许可证)

# 构建开发环境
成功编译Ontology需要以下准备：

* Golang版本在1.9及以上
* 安装第三方包管理工具glide
* 正确的Go语言开发环境
* Golang所支持的操作系统

# 部署及测试
## 获取ontology
### 从源码获取
克隆Ontology仓库到$GOPATH/src目录
```shell
$ git clone https://github.com/ontio/Ontology.git
```

用第三方包管理工具glide拉取依赖库

````shell
$ cd Ontology
$ glide install
````

用make编译源码

```shell
$ make
```

成功编译后会生成两个可以执行程序

* `ontology`: 节点程序
* `nodectl`: 以命令行方式提供的节点控制程序

### 从release获取
//TODO 将和release版本同步更新

## 创建ONT钱包文件
创建钱包文件
    - 通过命令行程序，在每个主机上分别创建节点运行所需的钱包文件wallet.dat 
      
        `$ ./nodectl wallet -c -p password` 

        注：通过-p参数设置钱包密码

## 服务器部署
成功运行Ontology可以通过以下两种方式进行部署

* 单机部署
* 多机部署
 * 在公共测试网上部署节点

### 单机部署配置

在单机上创建一个目录，在目录下存放以下文件：
- 默认配置文件`config.json`
- 节点程序`ontology`
- 节点控制程序`nodectl`
- 钱包文件`wallet.dat`
把source根目录下的config-solo.config配置文件的内容复制到config.json内，然后启动节点即可。

单机配置的例子如下：
- 目录结构
```shell
$ tree
└── node
    ├── config.json
    ├── ontology
    ├── nodectl
    └── wallet.dat
```

### 多机部署配置

网络环境下，最少需要4个节点（共识节点）完成部署。
我们可以通过修改默认的配置文件`config.json`进行快速部署。

1. 将相关文件复制到目标主机，包括：
    - 默认配置文件`config.json`
    - 节点程序`ontology`
    - 节点控制程序`nodectl`

2. 设置每个节点网络连接的端口号（推荐不做修改，使用默认端口配置）
    - `NodePort`为的P2P连接端口号（默认20338）
    - `HttpJsonPort`和`HttpLocalPort`为RPC端口号（默认为20336，20337）

3. 种子节点配置
    - 在4个主机中选出至少一个做种子节点，并将种子节点地址分别填写到每个配置文件的`SeelList`中，格式为`种子节点IP地址 + 种子节点NodePort`

4. 创建钱包文件
    - 通过命令行程序，在每个主机上分别创建节点运行所需的钱包文件wallet.dat 
      
        `$ ./nodectl wallet -c -p password` 

        注：通过-p参数设置钱包密码

5. 记账人配置
    - 为每个节点创建钱包时会显示钱包的公钥信息，将所有节点的公钥信息分别填写到每个节点的配置文件的`Bookkeepers`项中
    
        注：每个节点的钱包公钥信息也可以通过命令行程序查看：
    
        `$ ./nodectl wallet -l -p password` 


多机部署配置完成，每个节点目录结构如下

```shell
$ ls
config.json ontology nodectl wallet.dat
```

一个配置文件片段如下, 可以参考根目录下的config.json文件。

### 在公共测试网上部署节点
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
    "BookKeepers": [
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

了解更多请运行 `./nodectl --h`.

# 简单示例
[请看这里](https://github.com/ontio/documentation/tree/master/smart-contract-tutorial)

# 贡献代码

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

# 开源社区

## 网站

- http://ont.io/

# 许可证

Ontology遵守GNU Lesser General Public License, 版本3.0。 详细信息请查看项目根目录下的LICENSE文件。
