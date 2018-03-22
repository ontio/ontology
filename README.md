# Ontology

 Ontology is a decentralized distributed network protocol based on blockchain technology. Through peer-to-peer network.

## Highlight Features

 *  Scalable Lightweight Universal Smart Contract
 *  Crosschain Interactive Protocol
 *  Quantum-Resistant Cryptography (optional module)
 *  China National Crypto Standard (optional module)
 *  High Optimization of TPS
 *  P2P Link Layer Encryption
 *  Multiple Consensus Algorithm Support (VBFT/DBFT/RBFT/SBFT)
 *  Configurable Block Generation Time

# Building
The requirements to build Ontology are:
 *  Go version 1.9 or later
 *  Glide (a third-party package management tool)
 *  Properly configured Go environment
 
Clone the Ontology repository into the appropriate $GOPATH/src directory.


```shell
$ git clone https://github.com/dappledger/Ontology.git

```

Fetch the dependent third-party packages with glide.


````shell
$ cd Ontology
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
 
To run Ontology successfully, at least 4 nodes are required. The four nodes can be deployed in the following two way:

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
 
 Here's an snippet for configuration, note that `10.0.1.100` and `10.0.1.101` are public seed node's addresses:
 
 
```shell
$ cat config.json
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

## Configurations for single-host deployment

 Copy the executable file `node`, `nodectl` and configuration file `config.json` to 4 different directories on the single host. Then change each `config.json` file as following.
 *  The SeedList section should be same in all `config.json`.
 *  For the seed node, the `NodePort` is the same with the port in `SeedList` part.
 *  For each non-seed node, the `NodePort` should have different ports.
 *  Also make sure that the `HttpJsonPort` and `HttpLocalPort` of each node do not conflict with those of the current host.
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
## Getting Started

Start the seed node program first and then other nodes. Just run:

```shell
$ ./node
$ - input you wallet password
```

## Testing Ontology in an open environment
 
//TODO
add later.

# Contributing

Can I contribute patches to Ontology project?

Yes! Please open a pull request with signed-off commits. We appreciate your help!

You can also send your patches as emails to the developer mailing list.
Please join the Ontology mailing list or forum and talk to us about it.

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

## Site

* http://ont.io/

# License

The Ontology library (i.e. all code outside of the cmd directory) is licensed under the GNU Lesser General Public License v3.0, also included in our repository in the License file.