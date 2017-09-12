[![Build Status](https://travis-ci.org/DNAProject/DNA.svg?branch=master)](https://travis-ci.org/DNAProject/DNA)

# DNA (Distributed Networks Architecture)

DNA是go语言实现的基于区块链技术的去中心化的分布式网络协议。可以用来数字化资产和金融相关业务包括资产注册，发行，转账等。

## 特性

* 可扩展的轻量级通用智能合约
* 跨链交互协议（进行中）
* 抗量子密码算法 (可选择模块)
* 中国商用密码算法 (可选择模块)
* 高度优化的交易处理速度
* 基于IPFS的分布式存储和文件共享解决方案
* 节点访问权限控制
* P2P连接链路加密
* 多种共识算法支持 (DBFT/RBFT/SBFT)
* 可配置区块生成时间
* 可配置电子货币模型
* 可配置的分区共识(进行中)

# 编译
成功编译DNA需要以下准备：

* Go版本在1.8及以上
* 安装第三方包管理工具glide
* 正确的Go语言开发环境

克隆DNA仓库到$GOPATH/src目录


```shell
$ git clone https://github.com/DNAProject/DNA.git
```

用第三方包管理工具glide拉取依赖库


````shell
$ cd DNA
$ glide install
````

用make编译源码

```shell
$ make
```

成功编译后会生成两个可以执行程序

* `node`: 节点程序
* `nodectl`: 以命令行方式提供的节点控制程序

# 部署

成功运行DNA需要至少4个节点，可以通过两种方式进行部署

* 多机部署
* 单机部署

## 多机部署配置

我们可以通过修改默认的配置文件`config.json`进行快速部署。

1. 将相关文件复制到目标主机，包括：
    - 默认配置文件`config.json`
    - 节点程序`node`
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
    - 为每个节点创建钱包时会显示钱包的公钥信息，将所有节点的公钥信息分别填写到每个节点的配置文件的`BookKeepers`项中
    
        注：每个节点的钱包公钥信息也可以通过命令行程序查看：
    
        `$ ./nodectl wallet -l -p password` 


多机部署配置完成，每个节点目录结构如下

```shell
$ ls
config.json node nodectl wallet.dat
```

一个配置文件片段如下, 其中10.0.1.100、10.0.1.101等都是种子节点地址:
```shell
$ cat config.json
    ...
    "SeedList": [
      "10.0.1.100:10338",
      "10.0.1.101:10338",
      "10.0.1.102:10338"
    ],
    "BookKeepers": [
      "0322cfdb6a20401c2e44ede40b5282b2925fcff21cdc3814d782fd26026f1d023d",
      "02b639c019537839ba30b7c8c0396095da8838993492c07fe6ca11a5cf7b8fd2ca",
      "032c842494feba4e3dec3b9b7d9ad080ce63c81a41f7d79d2bbb5d499d16322907",
      "03d36828a99547184452276116f1b5171861931ff439a6da2316fddf1f3f428850"
    ],
    "HttpInfoPort": 10333,
    "HttpInfoStart": true,    
    "HttpRestPort": 10334,
    "HttpWsPort": 10335,
    "HttpJsonPort": 10336,
    "HttpLocalPort": 10337,
    "NoticeServerUrl":"",
    "OauthServerUrl":"",
    "NodePort": 10338,
    ...
```
## 单机部署配置

在单机上创建4个不同的目录，类似多机部署的方法分别在每个目录下存放以下文件：
- 默认配置文件`config.json`
- 节点程序`node`
- 节点控制程序`nodectl`
- 钱包文件`wallet.dat`
与多机配置不同的是，需要保证本机上端口不冲突, 请使用者自行修改个端口值。

单机配置的例子如下：
- 目录结构
```shell
$ tree
├── node1
│   ├── config.json
│   ├── node
│   ├── nodectl
│   └── wallet.dat
├── node2
│   ├── config.json
│   ├── node
│   ├── nodectl
│   └── wallet.dat
├── node3
│   ├── config.json
│   ├── node
│   ├── nodectl
│   └── wallet.dat
└── node4
    ├── config.json
    ├── node
    ├── nodectl
    └── wallet.dat
```
- 配置文件参考
```shell
$ cat node[1234]/config.json
    ...
    "SeedList": [
      "10.0.1.100:10338",
      "10.0.1.100:20338",
      "10.0.1.100:30338",
      "10.0.1.100:40338"
    ],
    "BookKeepers": [
      "0322cfdb6a20401c2e44ede40b5282b2925fcff21cdc3814d782fd26026f1d023d",
      "02b639c019537839ba30b7c8c0396095da8838993492c07fe6ca11a5cf7b8fd2ca",
      "032c842494feba4e3dec3b9b7d9ad080ce63c81a41f7d79d2bbb5d499d16322907",
      "03d36828a99547184452276116f1b5171861931ff439a6da2316fddf1f3f428850"
    ],
    "HttpInfoPort": 10333,
    "HttpInfoStart": true,    
    "HttpRestPort": 10334,
    "HttpWsPort": 10335,
    "HttpJsonPort": 10336,
    "HttpLocalPort": 10337,
    "NoticeServerUrl":"",
    "OauthServerUrl":"",
    "NodePort": 10338,
    ...

    "SeedList": [
      "10.0.1.100:10338",
      "10.0.1.100:20338",
      "10.0.1.100:30338",
      "10.0.1.100:40338"
    ],
    "BookKeepers": [
      "0322cfdb6a20401c2e44ede40b5282b2925fcff21cdc3814d782fd26026f1d023d",
      "02b639c019537839ba30b7c8c0396095da8838993492c07fe6ca11a5cf7b8fd2ca",
      "032c842494feba4e3dec3b9b7d9ad080ce63c81a41f7d79d2bbb5d499d16322907",
      "03d36828a99547184452276116f1b5171861931ff439a6da2316fddf1f3f428850"
    ],
    "HttpInfoPort": 20333,
    "HttpInfoStart": true,    
    "HttpRestPort": 20334,
    "HttpWsPort": 20335,
    "HttpJsonPort": 20336,
    "HttpLocalPort": 20337,
    "NoticeServerUrl":"",
    "OauthServerUrl":"",
    "NodePort": 20338,
    ...

    "SeedList": [
      "10.0.1.100:10338",
      "10.0.1.100:20338",
      "10.0.1.100:30338",
      "10.0.1.100:40338"
    ],
    "BookKeepers": [
      "0322cfdb6a20401c2e44ede40b5282b2925fcff21cdc3814d782fd26026f1d023d",
      "02b639c019537839ba30b7c8c0396095da8838993492c07fe6ca11a5cf7b8fd2ca",
      "032c842494feba4e3dec3b9b7d9ad080ce63c81a41f7d79d2bbb5d499d16322907",
      "03d36828a99547184452276116f1b5171861931ff439a6da2316fddf1f3f428850"
    ],
    "HttpInfoPort": 30333,
    "HttpInfoStart": true,    
    "HttpRestPort": 30334,
    "HttpWsPort": 30335,
    "HttpJsonPort": 30336,
    "HttpLocalPort": 30337,
    "NoticeServerUrl":"",
    "OauthServerUrl":"",
    "NodePort": 30338,
    ...

    "SeedList": [
      "10.0.1.100:10338",
      "10.0.1.100:20338",
      "10.0.1.100:30338",
      "10.0.1.100:40338"
    ],
    "BookKeepers": [
      "0322cfdb6a20401c2e44ede40b5282b2925fcff21cdc3814d782fd26026f1d023d",
      "02b639c019537839ba30b7c8c0396095da8838993492c07fe6ca11a5cf7b8fd2ca",
      "032c842494feba4e3dec3b9b7d9ad080ce63c81a41f7d79d2bbb5d499d16322907",
      "03d36828a99547184452276116f1b5171861931ff439a6da2316fddf1f3f428850"
    ],
    "HttpInfoPort": 40333,
    "HttpInfoStart": true,    
    "HttpRestPort": 40334,
    "HttpWsPort": 40335,
    "HttpJsonPort": 40336,
    "HttpLocalPort": 40337,
    "NoticeServerUrl":"",
    "OauthServerUrl":"",
    "NodePort": 40338,
    ...
    
```

## 运行
以任意顺序运行每个节点node程序，并在出现`Password:`提示后输入节点的钱包密码

```shell
$ ./node
$ - 输入你的钱包口令
```

## 测试环境

我们在云上部署了DNA供大家使用

主要功能包括：
1. 区块链相关信息查询
    - 区块信息
    - 交易信息
    - 节点信息
2. 资产操作
    - 注册资产
    - 发型资产
    - 转账
3. 测试交易发送

使用方式参见：

[forum.DNAProject.com/DNA节点控制工具](https://forum.dnaproject.org/t/dna-nodectl/57)

可用节点如下：
```
IP               PORT
----------------------
139.219.108.92:  10336
40.125.205.244:  10336
139.219.105.20:  10336
139.219.102.76:  10336
139.219.99.237:  10336
139.219.99.101:  10336
```

注：以上环境仅供测试使用，数据可能丢失或重置，我们不保证测试数据安全，请用户注意备份数据。

# 贡献代码

请您以签过名的commit发送pull request请求，我们期待您的加入！
您也可以通过邮件的方式发送你的代码到开发者邮件列表，欢迎加入DNA邮件列表和开发者论坛。

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

## 邮件列表

我们为开发者提供了一下邮件列表

- OnchainDNA@googlegroups.com

可以通过两种方式订阅并参与讨论

- 发送任何内容到邮箱地址 OnchainDNA+subscribe@googlegroups.com

- 登录 https://groups.google.com/forum/#!forum/OnchainDNA 


## 网站

- https://www.DNAproject.org

## 论坛

- https://forum.DNAproject.org

## Wiki

- https://wiki.DNAproject.org

# 许可证

DNA遵守Apache License, 版本2.0。 详细信息请查看项目根目录下的LICENSE文件。
