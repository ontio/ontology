[![Build Status](https://travis-ci.org/DNAProject/DNA.svg?branch=master)](https://travis-ci.org/DNAProject/DNA)

# DNA (Distributed Networks Architecture)

DNA is the golang implementation of a decentralized and distributed network protocol which is based on blockchain technology. It can be used for digitalize assets or shares and accomplish some financial business through peer-to-peer network such as registration, issuing, making transactions, settlement, payment and notary, etc.

## Highlight Features

* Extendable Generalduty Lightweight Smart Contract
* Crosschain Interactive Protocol
* Quantum-Resistant Cryptography (optional module)
* China National Crypto Standard (optional module)
* Zero Configuration Kickoff Networks (Ongoing)
* High Optimization of TPS
* IPFS Dentralizaed Storing & Sharing files solution integration (TBD)
* P2P link layer encryption, node access privilege controlling
* Multiple Consensus Algorithm(DBFT/RBFT/SBFT) support
* Telescopic Block Generation Time
* Digtal Asset Management
* Configurable Economic Incentive
* Configable sharding Consensus
* Configable Policy Management

# Building
The requirements to build DNA are:

* Go version 1.8 or later is required
* glide (third-party package management tool) is required
* A properly configured go environment


Clone the DNA repo into appropriate $GOPATH/src directory


```shell
$ git clone https://github.com/DNAProject/DNA.git

```

Fetch the dependent third-party packages with glide.


````shell
$ cd DNA
$ glide install
````

Build the source with make

```shell
$ make
```

After building the source code, you could see two executable programs you may need:

* `node`: the node program
* `nodectl`: command line tool for node control

Follow the precedures in Deployment section to give them a shot!


# Deployment

To run DNA node regularly, at least 4 nodes are necessary. We provides two ways to deploy the 4 nodes on:

* multi-hosts
* single-host

## Configurations for multi-hosts deployment

We can do a quick multi-hosts deployment by changing default configuration file `config.json`. Change the IP address in `SeedList` section to the seed node's IP address, then copy the changed file to hosts that you will run on.

On each host, put the executable program `node`, `nodectl` and the configuration file `config.json` into same directory. Like :

```shell
$ ls
config.json node nodectl

```
For each node, also needs a `wallet.dat` to run. The quick way to generate wallets is trying to run `./nodectl wallet -c -p YourPassword` on each host. 

Then, change the `BookKeepers` field to 4 nodes's wallet public keys, which you can got them from the last command's echo. The public key sequence is not matter. 

Now all configurations are completed.

Here's an snippet for configuration, note that `35.189.182.223` and `35.189.166.234` are two public seed node's addresses:
 
```shell
$ cat config.json
	...
    "SeedList": [
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

Copy the executable file `node`, `nodectl` and configuration file `config.json` to 4 different directories on single host. Then change each `config.json` file as following.

* The `SeedList` section should be same in all `config.json`.
* For the seed node, the `NodePort` is same with the port in `SeedList` part.
* For each non-seed node, the `NodePort` should have different port.
* Also need to make sure the `HttpJsonPort` and `HttpLocalPort` for each node is not conflict on current host.

After changed the configuration file, we also need to generate wallet for each node and field the `BookKeepers` with 4 nodes's wallet public keys. Please follow the steps in multi-hosts deployment section above.

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

## Getting Started

Start the seed node program firstly then other nodes. Just run:

```shell
$ ./node
$ - input you wallet password
```

## Testing DNA in an open environment

We also provide an open testing environment, it suppots below operation:

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

Some other avaliable nodes for testing:
```
IP               PORT
----------------------
35.189.182.223:  10336
35.189.182.223:  20336
35.189.166.234:  30336
35.189.166.234:  40336
```

`Notice: Above nodes intended be used for public testing only, the data saved on the testing chain maybe reset at anytime. Keep in mind to backup the data by youself to avoid data losting.`

# Contributing

Can I contribute patches to DNA project?

Yes! Please open a pull request with signed-off commits. We appreciate your help!

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

# Community

## Mailing list

We have a mailing list for developers:

* OnchainDNA@googlegroups.com

We provide two ways to subscribe:

* By sending any contents to email OnchainDNA+subscribe@googlegroups.com with any contents

* By signing in https://groups.google.com/forum/#!forum/OnchainDNA


## Forum

* https://www.DNAproject.org

## Wiki

* http://wiki.DNAproject.org

# License

DNA is licensed under the Apache License, Version 2.0. See LICENSE for the full license text.
