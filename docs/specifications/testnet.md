
<h1 align="center">Ontology </h1>
<h4 align="center">Version 0.7.0 </h4>

[![GoDoc](https://godoc.org/github.com/ontio/ontology?status.svg)](https://godoc.org/github.com/ontio/ontology)
[![Go Report Card](https://goreportcard.com/badge/github.com/ontio/ontology)](https://goreportcard.com/report/github.com/ontio/ontology)
[![Travis](https://travis-ci.org/ontio/ontology.svg?branch=master)](https://travis-ci.org/ontio/ontology)
[![Gitter](https://badges.gitter.im/Join%20Chat.svg)](https://gitter.im/ontio/ontology?utm_source=badge&utm_medium=badge&utm_campaign=pr-badge)


English | [中文](testnet_CN.md) 

# Server deployment

To run Ontology successfully,  nodes can be deployed by two ways:

- Single-host deployment
- Multi-hosts deployment
  - Deploy nodes on the public test network

### Single-host deployment configuration

Create a directory on the host and store the following files in the directory:

- Default configuration file `config.json`
- Node program + Node control program  `ontology`
- Wallet file`wallet.dat`, copy the contents of the configuration file config-solo.config in the root directory to config.json and start the node.
- Edit the config.json file and replace the bookkeeper entries with the public key of your wallet (created above). Use `$ ./ontology wallet show --name=wallet.dat` to get your public key.

Here's a example of single-host configuration:

- Directory structure
```shell
$ tree
└── ontology
    ├── config.json
    ├── ontology
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

2. Set the network connection port number for each node (recommend using the default port configuration, instead of modifying)

   - `NodePort`is P2P connection port number (default: 20338)
   - `HttpJsonPort` and `HttpLocalPort` are RPC port numbers (default: 20336, 20337)

3. Seed nodes configuration

   - Select at least one seed node out of 4 hosts and fill the seed node address into the `SeelList` of each configuration file. The format is `Seed node IP address + Seed node NodePort`.

4. Create wallet file

   - Through command line program, on each host create wallet wallet.dat needed for node implementation.

     `$ ./ontology wallet create --name=wallet.dat`

     Note: Set wallet password by parameter -p.

5. Bookkeepers configuration

   - While creating a wallet for each node, the public key information of the wallet will be displayed. Fill in the public key information of all nodes in the `Bookkeepers` field of each node's configuration file.

     Note: The public key information for each node's wallet can also be viewed via the command line program:

     `$ ./ontology wallet show --name=wallet.dat`

Now multi-host configuration is completed, directory structure of each node is as follows:

```
$ ls
config.json ontology wallet.dat
```

A configuration file fragment is as follows, you refer to the config.json file in the root directory.

### Deploy nodes on public test network (default config)

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

Run `./ontology --help` for details.

# Examples
## Contract
[Smart contract guide](https://github.com/ontio/documentation/tree/master/smart-contract-tutorial)

## ONT transfer sample
  contract:contract address； - from: transfer from； - to: transfer to； - value: amount；
```shell
  .\ontology asset transfer --caddr=ff00000000000000000000000000000000000001 --value=500 --from  TA6nAAdX77wcsAnuBQxG61zXg3vJUAPpgk  --to TA6Hsjww86b9KBbXFyKEayMcVVafoTGH4K  --password=xxx
```
