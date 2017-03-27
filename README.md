# DNA (Distributed Networks Architecture)

DNA (Distributed Networks Architecture) is the Golang implementation of a decentralized and distributed network protocol which is based on blockchain technology.
It can be used for digitalize assets or shares, and accomplish some financial business through peer-to-peer network such as registration and issuing,
Make transactions, settlement and payment. notary etc

It compatibles with antsahares[link] and targets to be more open source community friendly and better performance

DNA blockchain can be found at https://www.DNAdev.org

# HighLight Features

* 轻量级可扩展易编写智能合约［链接］
* 跨链协议crosschain compatible ［链接］
* 抗量子攻击 (可选组建)
* 国密算法支持SM系列支持 （可选组建）
* 零配置组网出块   （Ongoing）
* High Performance
* 分布式数据存储IPFS平滑集成 (TBD)
* 链接通道加密，CA节点认证可控
* 可插拔共识模块 DBFT/RBFT/SBFT and more
* 动态区块生成时间  （Telescopic block generate time)
* 数字资产管理模块  （Digtal Asset Management)
* 可配置化经济激励模型  （Configable Economic incentive)
* 可配置分区共识机制  （Configable sharding consensus)
* 可配置策略控制  （Configable Policy Management)
* 更多持续扩展...


# Contributing

Can I contribute patches to DNA project?

Yes! Please join the DNA mail list or forum and talk to us about it.

## Mailing list

There is a mailing list for developers: OnchainDNA@googlegroups.com

### How to Subscribe

* Send email to OnchainDNA+subscribe@googlegroups.com with any content or
* Go to https://groups.google.com/forum/#!forum/OnchainDNA to subscribe



## Dev Forum
* Forum https://www.DNAdevelopers.org
* Google group at https://groups.google.com/forum/#!forum/OnchainDNA
* Our user forum is at http://DNAdev.org/user-forum/


If you want to contribute code, please open a pull request with signed-off
commits at https://github.com/DNAdev/DNAChain/pulls
(alternatively, you can also send your patches as emails to the developer
mailing lsit).

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

where that header line really should be meaningful, and really should be
just one line.  That header line is what is shown by tools like gitk and
shortlog, and should summarize the change in one readable line of text,
independently of the longer explanation. Please use verbs in the
imperative in the commit message, as in "Fix bug that...", "Add
file/feature ...", or "Make Subsurface..."


# Basic Usage

Install and start from the desktop, or you can run it locally from the
build directory:

$ ./node

The official Windows installers are both cross-built on Linux & Mac.

You can give a data file as command line argument, or (once you have
set this up in the Preferences) Subsurface picks a default file for
you when started from the desktop or without an argument.

If you have a dive computer supported by libdivecomputer, you can just
select "Import from Divecomputer" from the "Import" menu, select which
dive computer you have (and where it is connected if you need to), and
hit "OK".

The latest list of supported dive computers can be found in the file
SupportedDivecomputers.txt.

Much more detailed end user instructions can be found from user guider
[Link]. When building from source this is also available as
Documentation/user-manual.md. The documentation for the latest release
is also available on-line http://www.DNAdev.org/documentation/

## Setup testbed at one host

You can get the sources to the latest development version from the git
repository:

git clone git://github.com/DNAdev/DNAChain

You can also fork the repository and browse the sources at the same site,
simply using https://github.com/DNAdev/DNAChain

If you want the latest release (instead of the bleeding edge
development version) you can either get this via git or the release tar
ball. After cloning run the following command:

git checkout v0.1  (or whatever the last release is)

or download a tar ball from:

http://DNAdev.org/downloads/DNAchain-0.1tgz

Detailed build instructions can be found in the INSTALL file.


单机建四节点测试简要步骤：
1: 创建四个不同的目录 IE：
./test1
./test2
./test3
./test3

2: 在各自目录下面拷贝缺省配置文件config.json IE:
./test1/config.json

3: 修改各自的config.json文件中的种子节点端口号 IE:
"SeedList": [
             "10.0.1.26:20338"
             ],
"HttpJsonPort": 20342,
“HttpLocalPort": 20343,
"NodePort": 20344,

＊种子节点端口号必须为四个节点中的一个
＊如果节点运行在同一台机器上，端口号必须不同

4: 在各自目录下运行 node 程序

5: 用nodectl去查看各个节点运行状态和发送测试用例 IE：
$ ./nodectl test -tx


### Setup nodes on different host

启动顺序同单机启动多节点，无需修改节点端口号

# License

Apache License Version 2.0


[![Go Report Card](https://goreportcard.com/badge/github.com/dreamfly281/GoOnchain)](https://goreportcard.com/report/github.com/dreamfly281/GoOnchain)
[![Build Status](https://travis-ci.org/dreamfly281/GoOnchain.png)](https://travis-ci.org/dreamfly281/GoOnchain)
