# DNA


## Distributed Networks Architecture
DNA is the golang implementation of a decentralized and distributed network protocol which is based on blockchain technology. It can be used for digitalize assets or shares and accomplish some financial business through peer-to-peer network such as registration, issuing, making transactions, settlement, payment and notary, etc.

## Features

* 可扩展通用智能合约平台 / Extendable Generalduty Lightweight Smart Contract
* 跨链协议 / Crosschain Interactive Protocol
* 抗量子攻击 / Quantum-Resistant Cryptography (optional module)
* 国密算法支持SM系列支持 / China National Crypto Standard (optional module)
* 零配置组网出块 / Zero Configuration Kickoff Networks (Ongoing)
* 高度优化交易处理速度 ／ High Optimization of TPS
* 分布式数据存储IPFS集成方案 ／ IPFS Dentralizaed Storing & Sharing files solution integration (TBD)
* 链接通道加密，CA节点认证可控 / P2P link layer encryption, node access privilege controlling
* 可插拔共识模块 / Multiple Consensus Algorithm(DBFT/RBFT/SBFT) support
* 动态区块生成时间 / Telescopic Block Generation Time
* 数字资产管理 / Digtal Asset Management
* 可配置化经济激励模型 / Configurable Economic Incentive
* 可配置分区共识机制 / Configable sharding Consensus
* 可配置策略控制 / Configable Policy Management

## Status
[![Build Status](https://travis-ci.org/DNAProject/DNA.svg?branch=master)](https://travis-ci.org/DNAProject/DNA)

# Building
The requirements to build DNA are:

* Go version 1.7 and later is required.
* glide (third-party package management tool) is required.
* A properly configured go environment.

Clone the DNA repo into appropriate GOPATH/src directory

```shell
$ git clone https://github.com/DNAProject/DNA.git

```

Fetch the dependent third-party packages with glide.


````shell
$ cd DNA; glide up;
````

Build the source with make

```shell
$ make
```

After building the source code, you could see two executable programs you may need:

* `node` : the node program
* `nodectl` : command line tool for node control

Follow the precedures in Depolyment section to give them a shot!


# Depolyment

To run DNA node regularly, at least 4 nodes are necessary. We provides two ways to deploy the 4 nodes on:

* multi-host
* single-host

## Configurations for multi-host depolyment

We can do a quick multi-host deployment by changing default configuration file `config/config.json`. Change the IP address in `SeedList` section to the seed node's IP address, then copy the changed file to hosts that you will run on.

On each host, put the executable program `node` and the configuration file `config.json` into same directory. Like :

```shell
$ ls
config.json node

```

Here's an snippet for configuration, note that 10.0.0.100 is seed node's address:

```shell
$ cat config.json
	...
"SeedList": [
      "10.0.1.100:20338"
    ],
	...
    "NodePort": 20338,
	...
```

## Configurations for single-host depolyment

Copy the executable file `node` and configuration file `config.json` to 4 different directories on single host. Then change each `config.json` file as following. 

* The `SeedList` section should be same in all `config.json`. 
* For the seed node, the `NodePort` is same with the port in `SeedList` part.
* For each non-seed node, the `NodePort` should have different port.
* Also need to make sure the `HttpJsonPort` and `HttpLocalPort` for each node is not conflict on current host.

Here's an example:

```shell
# directory structure #
$ tree
├── node1
│   ├── config.json
│   └── node
├── node2
│   ├── config.json
│   └── node
├── node3
│   ├── config.json
│   └── node
└── node4
    ├── config.json
    └── node
```

```shell
# configuration snippets #
$ cat node[1234]/config.json
"SeedList": [
      "10.0.1.1:10338"
    ],
    "HttpJsonPort": 10336,
    "HttpLocalPort": 10337,
    "NodePort": 10338,
    ...

"SeedList": [
      "10.0.1.1:10338"
    ],
    "HttpJsonPort": 20336,
    "HttpLocalPort": 20337,
    "NodePort": 20338,
    ...

"SeedList": [
      "10.0.1.1:10338"
    ],
    "HttpJsonPort": 30336,
    "HttpLocalPort": 30337,
    "NodePort": 30338,
    ...

"SeedList": [
      "10.0.1.1:10338"
    ],
    "HttpJsonPort": 40336,
    "HttpLocalPort": 40337,
    "NodePort": 40338,
    ...
```

## Run

Execute the seed node program first then other nodes. Just run:

```shell
$ ./node
```


# Contributing

>Can I contribute patches to DNA project?

>Yes! Please open a pull request with signed-off commits. We appreciate your help!

You can also send your patches as emails to the developer mailing list.
Please join the DNA mailing list or forum and talk to us about it.

Either way, if you don't sign off your patches, we will not accept them.
This means adding a line that says "Signed-off-by: Name <email>" at the
end of each commit, indicating that you wrote the code and have the right
to pass it on as an open source patch.

See: http://developercertificate.org/

Also, please write good git commit messages.  A good commit message
looks like this:

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

## Mailing list

We have a mailing list for developers: 

* OnchainDNA@googlegroups.com

We provide two ways to subscribe:

* By sending any contents to email OnchainDNA+subscribe@googlegroups.com with any contents

* By signing in https://groups.google.com/forum/#!forum/OnchainDNA

# Community
Most technical topics could be found at our forum

* https://forum.DNAproject.org

# License

DNA is licensed under the Apache License, Version 2.0. See LICENSE for the full license text.
