[![Build Status](https://travis-ci.org/DNAProject/DNA.svg?branch=master)](https://travis-ci.org/DNAProject/DNA)

# DNA (Distributed Networks Architecture)

 DNA is a decentralized distributed network protocol based on blockchain technology and is implemented in Golang. Through peer-to-peer network, DNA can be used to digitize assets and provide financial service, including asset registration, issuance, transfer, etc.

## Highlight Features

 *	Scalable Lightweight Universal Smart Contract
 *	Crosschain Interactive Protocol
 *	Quantum-Resistant Cryptography (optional module)
 *	China National Crypto Standard (optional module)
 *	High Optimization of TPS
 *	Distributed Storage and File Sharding Solutions Based on IPFS
 *	P2P Link Layer Encryption
 *	Node Access Control
 *	Multiple Consensus Algorithm Support (DBFT/RBFT/SBFT)
 *	Configurable Block Generation Time
 *	Configurable Digital Currency Incentive
 *	Configable Sharding Consensus (in progress)


# Building
The requirements to build DNA are:
 *	Go version 1.8 or later
 *	Glide (a third-party package management tool)
 *	Properly configured Go environment
 
Clone the DNA repository into the appropriate $GOPATH/src directory.


```shell
$ git clone https://github.com/DNAProject/DNA.git

```

Fetch the dependent third-party packages with glide.


````shell
$ cd DNA
$ glide install
````
Build the source code with make.

```shell
$ make
```

After building the source code, you should see two executable programs:

* `node`: the node program
* `nodectl`: command line tool for node control

Follow the procedures in Deployment section to give them a shot!


# Deployment
 
To run DNA successfully, at least 4 nodes are required. The four nodes can be deployed in the following two way:

* multi-hosts deployment
* single-host deployment

## Configurations for multi-hosts deployment

 We can do a quick multi-host deployment by modifying the default configuration file `config.json`. Change the IP address in `SeedList` section to the seed node's IP address, and then copy the changed file to the hosts that you will run on.
 On each host, put the executable program `node`, `nodectl` and the configuration file `config.json` into the same directory. Like :
 
```shell
$ ls
config.json node nodectl

```
 Each node also needs a `wallet.dat` to run. The quickest way to generate wallets is to run `./nodectl wallet -c -p YourPassword` on each host.
 Then, change the `BookKeepers` field to the 4 nodes' wallet public keys, which you can get from the last command's echo. The public key sequence does not matter.
 Now all configurations are completed.
 
 Here's an snippet for configuration, note that `35.189.182.223` and `35.189.166.234` are two public seed node's addresses:
 
 
```shell
$ cat config.json
	...
    "SeedList": [
      "35.189.182.223:10338",
      "35.189.182.223:20338",
      "35.189.166.234:30338"
    ],
	...
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
```

## Configurations for single-host deployment

 Copy the executable file `node`, `nodectl` and configuration file `config.json` to 4 different directories on the single host. Then change each `config.json` file as following.
 *	The SeedList section should be same in all `config.json`.
 *	For the seed node, the `NodePort` is the same with the port in `SeedList` part.
 *	For each non-seed node, the `NodePort` should have different ports.
 *	Also make sure that the `HttpJsonPort` and `HttpLocalPort` of each node do not conflict with those of the current host.
 After changing the configuration file, we also need to generate a wallet for each node and field the `BookKeepers` with the 4 nodes' wallet public keys. Please follow the steps in the multi-hosts deployment section above.
 Here's an example:


```shell
# directory structure #
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

```shell
# configuration snippets #
$ cat node[1234]/config.json
    "SeedList": [
      "35.189.182.223:10338",
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
      "35.189.182.223:10338",
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
      "35.189.182.223:10338",
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
      "35.189.182.223:10338",
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

## Getting Started

Start the seed node program first and then other nodes. Just run:

```shell
$ ./node
$ - input you wallet password
```

## Testing DNA in an open environment
 
 We also provide an open testing environment. It supports the operation below:

1. make some transactions :
```
./nodectl test -ip 35.189.182.223 -port 10336 -tx perf -num 10
```

2. register, issue, transfer assert :
```
./nodectl test -ip 35.189.182.223 -port 10336 -tx full
```

3. look up block's information :
```
./nodectl info -ip 35.189.182.223 -port 10336 -height 10
```

4. look up transaction's information :
```
./nodectl info -ip 35.189.182.223 -port 10336 -txhash d438896f07786b74281bc70259b0caaccb87460171104ea17473b5e802033a98
```

......

Run `./nodectl --h` for more details.

Some other available nodes for testing:
```
IP               PORT
----------------------
35.189.182.223:  10336
35.189.182.223:  20336
35.189.166.234:  30336
35.189.166.234:  40336
```
 
 `Notice: The nodes above are intended to be used for public testing only. The data saved on the testing chain maybe be reset at any time. Keep in mind to back up the data by yourself to avoid data loss.`

# Contributing

Can I contribute patches to DNA project?

Yes! Please open a pull request with signed-off commits. We appreciate your help!

You can also send your patches as emails to the developer mailing list.
Please join the DNA mailing list or forum and talk to us about it.

Either way, if you don't sign off your patches, we will not accept them.
This means adding a line that says "Signed-off-by: Name <email>" at the
end of each commit, indicating that you wrote the code and have the right
to pass it on as an open source patch.

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

# Community

## Mailing list

We have a mailing list for developers:

* OnchainDNA@googlegroups.com

We provide two ways to subscribe:

* Send any contents to the email OnchainDNA+subscribe@googlegroups.com

* Sign in https://groups.google.com/forum/#!forum/OnchainDNA


## Forum

* https://www.DNAproject.org

## Wiki

* http://wiki.DNAproject.org

# License

DNA is licensed under the Apache License, Version 2.0. See LICENSE for the full license text.
