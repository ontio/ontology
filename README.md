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

* Go version 1.7 or later is required
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

We can do a quick multi-hosts deployment by changing default configuration file `config/config.json`. Change the IP address in `SeedList` section to the seed node's IP address, then copy the changed file to hosts that you will run on.

On each host, put the executable program `node` and the configuration file `config.json` into same directory. Like :

```shell
$ ls
config.json node

```

We need to do is change the `MinerName` field to "c1", "c2", "c3" and "c4" respectively on each host. The name sequence is not matter.

Here's an snippet for configuration, note that 10.0.0.100 is seed node's address:

```shell
$ cat config.json
	...
"SeedList": [
      "10.0.1.100:20338"
    ],
	...
    "MinerName" : "c1"
    "NodePort": 20338,
	...
```

For each node, also needs a "wallet" to run. The quick way to generate wallets is trying to run the node program on a host, several wallets named with "wallet" prefix will be generated automatically. Then copy all of them to node program directory on other hosts.
Congratulations, all configurations are completed.

## Configurations for single-host deployment

Copy the executable file `node` and configuration file `config.json` to 4 different directories on single host. Then change each `config.json` file as following. 

* The `SeedList` section should be same in all `config.json`. 
* For the seed node, the `NodePort` is same with the port in `SeedList` part.
* For each non-seed node, the `NodePort` should have different port.
* Also need to make sure the `HttpJsonPort` and `HttpLocalPort` for each node is not conflict on current host.
* Each node should have different "MinerName" field, "c1", "c2", "c3" and "c4" respectively.

After changed the configuration file, we also need to generate wallet for each node. Please follow the steps in multi-hosts deployment section above.

Here's an example:

```shell
# directory structure #
$ tree
├── node1
│   ├── config.json
│   ├── node
│   └── wallet*
├── node2
│   ├── config.json
│   ├── node
│   └── wallet*
├── node3
│   ├── config.json
│   ├── node
│   └── wallet*
└── node4
    ├── config.json
    ├── node
    └── wallet*
```

```shell
# configuration snippets #
$ cat node[1234]/config.json
"SeedList": [
      "10.0.1.1:10338"
    ],
    "MinerName" : "c1"
    "HttpJsonPort": 10336,
    "HttpLocalPort": 10337,
    "NodePort": 10338,
    ...

"SeedList": [
      "10.0.1.1:10338"
    ],
    "MinerName" : "c2"
    "HttpJsonPort": 20336,
    "HttpLocalPort": 20337,
    "NodePort": 20338,
    ...

"SeedList": [
      "10.0.1.1:10338"
    ],
    "MinerName" : "c3"
    "HttpJsonPort": 30336,
    "HttpLocalPort": 30337,
    "NodePort": 30338,
    ...

"SeedList": [
      "10.0.1.1:10338"
    ],
    "MinerName" : "c4"
    "HttpJsonPort": 40336,
    "HttpLocalPort": 40337,
    "NodePort": 40338,
    ...
```

## Getting Started

Start the seed node program firstly then other nodes. Just run:

```shell
$ ./node
```

## Testing DNA in an open environment

We also provide an open testing environment, it suppots below operation:

1. make some transactions :
```
./nodectl test -ip 139.196.113.85 -port 10331 -tx perf -num 10
```

2. register, issue, transfer assert :
```
./nodectl test -ip 139.196.113.85 -port 10331 -tx full
```

3. look up block's information :
```
./nodectl info -ip 139.196.113.85 -port 10331 -height 10
```

4. look up transaction's information :
```
./nodectl info -ip 139.196.113.85 -port 10331 -txhash d438896f07786b74281bc70259b0caaccb87460171104ea17473b5e802033a98
```

......

Run `./nodectl --h` for more details.

Some other avaliable nodes for testing:
```
IP               PORT
----------------------
139.196.113.85:  10331
139.196.113.101: 10331
139.196.227.195: 10331
139.196.227.195: 20331
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
