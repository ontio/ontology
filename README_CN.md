[![Build Status](https://travis-ci.org/DNAProject/DNA.svg?branch=master)](https://travis-ci.org/DNAProject/DNA)

# DNA (Distributed Networks Architecture)

DNA是go语言实现的基于区块链技术的去中心化的分布式网络协议。可以用来数字化资产和金融相关业务包括资产注册，发行，转账等。

## 特性

* 可扩展的轻量级通用智能合约
* 跨链交互协议（进行中）
* 抗量子密码算法 (可选择模块)
* 中国商用密码算法 (可选择模块)
* 零知识证明网络 (进行中)
* 高度优化的交易处理速度
* 基于IPFS的分布式存储和文件共享解决方案
* P2P连接加密
* 节点访问权限控制
* 多种共识算法支持 (DBFT/RBFT/SBFT)
* 可配置区块生成时间
* 数字资产管理
* 可配置电子货币激励模型
* 可配置的分区共识(进行中)
* 可配置的策略管理机制

# 编译
成功编译DNA需要以下准备：

* Go版本在1.7及以上
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

一个配置文件片段如下, 其中10.0.0.100是种子节点地址:
```shell
$ cat config.json
    ...
    "SeedList": [
      "35.189.182.223:20338",
      "35.189.166.234:30338"
    ],
    "BookKeepers": [
      "03ad8f4a837f7a02adedcea920b30c5c99517aabc7d2695d93ac572b9c2106d4c2",
      "0293bafa2df4813ae999bf42f35a38bcb9ec26a252fd28dc0ccab56c671cf784e6",
      "02aec70e084e4e5d36ed2db54aa708a6bd095fbb663929850986a5ec22061e1be2",
      "02758623d16774f3c5535a305e65ea949343eab06888ee2e7633b4f3f9d78d506c"
    ],
	"HttpRestPort": 20334,
    "HttpJsonPort": 20336,
    "HttpLocalPort": 20337,
    "NodePort": 20338,
    ...

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
      "35.189.182.223:20338",
      "35.189.166.234:30338"
    ],
    "BookKeepers": [
      "03ad8f4a837f7a02adedcea920b30c5c99517aabc7d2695d93ac572b9c2106d4c2",
      "0293bafa2df4813ae999bf42f35a38bcb9ec26a252fd28dc0ccab56c671cf784e6",
      "02aec70e084e4e5d36ed2db54aa708a6bd095fbb663929850986a5ec22061e1be2",
      "02758623d16774f3c5535a305e65ea949343eab06888ee2e7633b4f3f9d78d506c"
    ],
    "HttpRestPort": 10334,
    "HttpJsonPort": 10336,
    "HttpLocalPort": 10337,
    "NodePort": 10338,
    ...

    "SeedList": [
      "35.189.182.223:20338",
      "35.189.166.234:30338"
    ],
    "BookKeepers": [
      "03ad8f4a837f7a02adedcea920b30c5c99517aabc7d2695d93ac572b9c2106d4c2",
      "0293bafa2df4813ae999bf42f35a38bcb9ec26a252fd28dc0ccab56c671cf784e6",
      "02aec70e084e4e5d36ed2db54aa708a6bd095fbb663929850986a5ec22061e1be2",
      "02758623d16774f3c5535a305e65ea949343eab06888ee2e7633b4f3f9d78d506c"
    ],
    "HttpRestPort": 20334,
    "HttpJsonPort": 20336,
    "HttpLocalPort": 20337,
    "NodePort": 20338,
    ...

    "SeedList": [
      "35.189.182.223:20338",
      "35.189.166.234:30338"
    ],
    "BookKeepers": [
      "03ad8f4a837f7a02adedcea920b30c5c99517aabc7d2695d93ac572b9c2106d4c2",
      "0293bafa2df4813ae999bf42f35a38bcb9ec26a252fd28dc0ccab56c671cf784e6",
      "02aec70e084e4e5d36ed2db54aa708a6bd095fbb663929850986a5ec22061e1be2",
      "02758623d16774f3c5535a305e65ea949343eab06888ee2e7633b4f3f9d78d506c"
    ],
    "HttpRestPort": 30334,
    "HttpJsonPort": 30336,
    "HttpLocalPort": 30337,
    "NodePort": 30338,
    ...

    "SeedList": [
      "35.189.182.223:20338",
      "35.189.166.234:30338"
    ],
    "BookKeepers": [
      "03ad8f4a837f7a02adedcea920b30c5c99517aabc7d2695d93ac572b9c2106d4c2",
      "0293bafa2df4813ae999bf42f35a38bcb9ec26a252fd28dc0ccab56c671cf784e6",
      "02aec70e084e4e5d36ed2db54aa708a6bd095fbb663929850986a5ec22061e1be2",
      "02758623d16774f3c5535a305e65ea949343eab06888ee2e7633b4f3f9d78d506c"
    ],
    "HttpRestPort": 40334,
    "HttpJsonPort": 40336,
    "HttpLocalPort": 40337,
    "NodePort": 40338,
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

[www.DNAProject.com/DNA节点控制工具](https://www.dnaproject.org/t/dna-nodectl/57)

可用节点如下：
```
IP               PORT
----------------------
35.189.182.223:  10336
35.189.182.223:  20336
35.189.166.234:  30336
35.189.166.234:  40336
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


## 论坛

- https://www.DNAproject.org

## Wiki

- http://wiki.DNAproject.org

# 许可证

DNA遵守Apache License, 版本2.0。 详细信息请查看项目根目录下的LICENSE文件。
