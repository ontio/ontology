
<h1 align="center">Ontology </h1>
<h4 align="center">Version 0.7.0 </h4>

[![GoDoc](https://godoc.org/github.com/ontio/ontology?status.svg)](https://godoc.org/github.com/ontio/ontology)
[![Go Report Card](https://goreportcard.com/badge/github.com/ontio/ontology)](https://goreportcard.com/report/github.com/ontio/ontology)
[![Travis](https://travis-ci.org/ontio/ontology.svg?branch=master)](https://travis-ci.org/ontio/ontology)
[![Gitter](https://badges.gitter.im/Join%20Chat.svg)](https://gitter.im/ontio/ontology?utm_source=badge&utm_medium=badge&utm_campaign=pr-badge)

English | [中文](README_CN.md)

Welcome to Ontology's source code library!

Ontology is dedicated to creating a modularized, freely configurable, interoperable cross-chain, high-performance, and horizontally scalable blockchain infrastructure system. Ontology makes deploying and invoking decentralized applications easier.

The code is currently alpha quality, but is in the process of rapid development. The master code may be unstable; stable versions can be downloaded in the release page.

The public test network is described below. We sincerely welcome and hope more developers join Ontology.

## Features

- Scalable lightweight universal smart contract
- Scalable WASM contract support
- Crosschain interactive protocol (processing)
- Multiple encryption algorithm support
- Highly optimized transaction processing speed
- P2P link layer encryption (optional module)
- Multiple consensus algorithm support (VBFT/DBFT/RBFT/SBFT/PoW)
- Quick block generation time


## Contents

* [Build development environment](#build-development-environment)
* [Deployment and test](#deployment-and-test)
	* [Get Ontology](#get-ontology)
		* [Get from source code](#get-from-source-code)
	* [Create ONT wallet file](#create-ont-wallet-file)
	* [Server deployment](#server-deployment)
		* [Single-host deployment configuration](#single-host-deployment-configuration)
		* [Multi-host deployment configuration](#multi-hosts-deployment-configuration)
		* [Deploy nodes on public test network](#deploy-nodes-on-public-test-network)
		* [Implement](#implement)
* [Examples](#examples)
* [Contributions](#contributions)
* [Open source community](#open-source-community)
	* [Site](#site)
	* [Developer Discord Group](#developer-discord-group)
* [License](#license)

# Build development environment

The requirements to build Ontology are:

- Golang version 1.9 or later
- Glide (a third party package management tool)
- Properly configured Go language environment
- Golang supported operating system

# Deployment and test
## Get Ontology
### Get from source code

Clone the Ontology repository into the appropriate $GOPATH/src/github.com/ontio directory.

```
$ git clone https://github.com/ontio/ontology.git
```
or
```
$ go get github.com/ontio/ontology
```
Fetch the dependent third party packages with glide.

```
$ cd $GOPATH/src/github.com/ontio/ontology
$ glide install
```

Build the source code with make.

```
$ make
```

After building the source code sucessfully, you should see two executable programs:

- `ontology`: the node program
- `nodectl`: command line program for node control

## Create ONT wallet file

## Create Ontology wallet
ONT supports multiple encryption methods for generating accounts, but can set a default in config.json such as SHA256withECDSA. 

Create wallet cmd:

```shell
$ ./nodectl wallet --create --name wallet.dat --password passwordtest
```

Note: Set wallet password by parameter -p.

To show the wallet info:

```shell
$ ./nodectl wallet --list account

public key:    1202021401156f187ec23ce631a489c3fa17f292171009c6c3162ef642406d3d09c74d
hex address:  018f0dcf09ec2f0040e6e8d7e54635dba40f7d63
base58 address:       TA7T3p6ikRG5s2pAaehUH2XvRCCzvsFmwE

$ ./nodectl wallet --list -b
ont: 248965536

* with -b cmd will show the ont amount this account have.
```

ONT supported crypto (<hash>with<dsa>):
 - SHA224withECDSA 
 - SHA256withECDSA
 - SHA384withECDSA
 - SHA512withECDSA
 - SHA3-224withECDSA
 - SHA3-256withECDSA
 - SHA3-384withECDSA
 - SHA3-512withECDSA
 - RIPEMD160withECDSA
 - SM3withSM2
 - SHA512withEdDSA

## Server deployment

To run Ontology successfully,  nodes can be deployed by two ways:

- Single-host deployment
- Multi-hosts deployment
  - Deploy nodes on the public test network

### Single-host deployment configuration

Create a directory on the host and store the following files in the directory:

- Default configuration file `config.json`
- Node program `ontology`
- Node control program `nodectl`
- Wallet file`wallet.dat`, copy the contents of the configuration file config-solo.config in the root directory to config.json and start the node.
- Edit the config.json file and replace the bookkeeper entries with the public key of your wallet (created above). Use `$ ./nodectl wallet -l -p password` to get your public key.

Here's a example of single-host configuration:

- Directory structure
```shell
$ tree
└── ontology
    ├── config.json
    ├── ontology
    ├── nodectl
    └── wallet.dat
```

Bookkeepers in the config.json file:
```
"Bookkeepers": [ "1202021c6750d2c5d99813997438cee0740b04a73e42664c444e778e001196eed96c9d" ],
```

### Multi-hosts deployment configuration

We can perform a quick deployment by modifying the default configuration file `config.json`.

1. Copy related file into target host, including:

   - Default configuration file`config.json`
   - Node program`ontology`
   - Node control program`nodectl`

2. Set the network connection port number for each node (recommend using the default port configuration, instead of modifying)

   - `NodePort`is P2P connection port number (default: 20338)
   - `HttpJsonPort` and `HttpLocalPort` are RPC port numbers (default: 20336, 20337)

3. Seed nodes configuration

   - Select at least one seed node out of 4 hosts and fill the seed node address into the `SeelList` of each configuration file. The format is `Seed node IP address + Seed node NodePort`.

4. Create wallet file

   - Through command line program, on each host create wallet wallet.dat needed for node implementation.

     `$ ./nodectl wallet -c -p password`

     Note: Set wallet password by parameter -p.

5. Bookkeepers configuration

   - While creating a wallet for each node, the public key information of the wallet will be displayed. Fill in the public key information of all nodes in the `Bookkeepers` field of each node's configuration file.

     Note: The public key information for each node's wallet can also be viewed via the command line program:

     `$ ./nodectl wallet -l -p password`

Now multi-host configuration is completed, directory structure of each node is as follows:

```
$ ls
config.json ontology nodectl wallet.dat
```

A configuration file fragment is as follows, you refer to the config.json file in the root directory.

### Deploy nodes on public test network

Start with the following configuration file to connect to the current ONT test network.

```
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
    "Bookkeepers": [
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

### Implement

Run each node program in any order and enter the node's wallet password after the `Password:` prompt appears.

```
$ ./ontology
$ - Input your wallet password
```

Run `./nodectl --h` for details.

# Examples
## Contract
[Smart contract guide](https://github.com/ontio/documentation/tree/master/smart-contract-tutorial)

## ONT transfer sample

```shell
  ./nodectl transfer --contract ff00000000000000000000000000000000000001 --value 10 --from 0181beb9cfba23c777421eaf57e357e0fc331cbf --to 01f3aecd2ba7a5b704fbd5bac673e141d5109e3e

  contract:contract address； - from: transfer from； - to: transfer to； - value: amount；
```

# Contributions

Please open a pull request with a signed commit. We appreciate your help! You can also send your code as emails to the developer mailing list. You're welcome to join the Ontology mailing list or developer forum.

Please provide detailed submission information when you want to contribute code for this project. The format is as follows:

Header line: explain the commit in one line (use the imperative).

Body of commit message is a few lines of text, explaining things  in more detail, possibly giving some background about the issue  being fixed, etc.

The body of the commit message can be several paragraphs. Please do proper word-wrap and keep columns shorter than 74 characters or so. That way "git log" will show things  nicely even when it is indented.

Make sure you explain your solution and why you are doing what you are  doing, as opposed to describing what you are doing. Reviewers and your  future self can read the patch, but might not understand why a  particular solution was implemented.

Reported-by: whoever-reported-it &
Signed-off-by: Your Name [youremail@yourhost.com](mailto:youremail@yourhost.com)

# Open source community
## Site

- <https://ont.io/>

## Developer Discord Group

- <https://discord.gg/4TQujHj/>

# License

The Ontology library is licensed under the GNU Lesser General Public License v3.0, read the LICENSE file in the root directory of the project for details.
