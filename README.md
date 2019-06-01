
<h1 align="center">Ontology</h1>
<h4 align="center">Version 1.6.0</h4>

[![GoDoc](https://godoc.org/github.com/ontio/ontology?status.svg)](https://godoc.org/github.com/ontio/ontology)
[![Go Report Card](https://goreportcard.com/badge/github.com/ontio/ontology)](https://goreportcard.com/report/github.com/ontio/ontology)
[![Travis](https://travis-ci.org/ontio/ontology.svg?branch=master)](https://travis-ci.org/ontio/ontology)
[![Discord](https://img.shields.io/discord/102860784329052160.svg)](https://discord.gg/gDkuCAq)

English | [中文](README_CN.md)

Welcome to the official source code repository for Ontology!

Ontology is a high-performance public blockchain project and distributed trust collaboration platform. It is highly customizable and suitable for all kinds of business requirements. The Ontology MainNet was launched on June 30th, 2018.

As a public blockchain project, Ontology is currently maintained by both the Ontology core tech team and community members who can all support you in development. There are many available tools for use for development - SDKs, the SmartX IDE, Ontology blockchain explorer and more.

New features are still being rapidly developed, therefore the master branch may be unstable. Stable versions can be found in the [releases section](https://github.com/ontio/ontology/releases).

- [Features](#features)
- [Build Development Environment](#build-development-environment)
- [Download Ontology](#download-ontology)
    - [Download Release](#download-release)
    - [Build from Source Code](#build-from-source-code)
- [Run Ontology](#run-ontology)
    - [MainNet Sync Node](#mainnet-sync-node)
    - [TestNet Sync Node](#testnet-sync-node)
    - [Local PrivateNet](#local-privatenet)
    - [Run with Docker](#run-in-docker)
- [Examples](#examples)
    - [ONT transfer sample](#ont-transfer-sample)
    - [Query transfer status sample](#query-transfer-status-sample)
    - [Query account balance sample](#query-account-balance-sample)
- [Contributions](#contributions)
- [Open source community](#open-source-community)
    - [Site](#site)
    - [Developer Discord Group](#developer-discord-group)
- [License](#license)

## Features

- Scalable lightweight universal smart contracts
- Scalable WASM contract support
- Cross-chain interactive protocol
- Multiple encryption algorithms supported
- Highly optimized transaction processing speed
- P2P link layer encryption (optional module)
- Multiple consensus algorithms supported (VBFT/DBFT/RBFT/SBFT/PoW)
- Quick block generation time (1-30 seconds)


## Build Development Environment
The requirements to build Ontology are:

- [Golang](https://golang.org/doc/install) version 1.9 or later
- [Glide](https://glide.sh) (a third party package management tool for Golang)

## Download Ontology

### Download Release
You can download a stable compiled version of the Ontology node software by either:

- Downloading the latest Ontology binary file with `curl https://dev.ont.io/ontology_install | sh`.
- Downloading a specific version from the [release section](https://github.com/ontio/ontology/releases).

### Build from Source Code
Alternatively, you build your own version directly from the source code. Note that the code in the `master` branch may not be stable.

1) Clone the Ontology repository into the appropriate `$GOPATH/src/github.com/ontio` directory.

```
$ git clone https://github.com/ontio/ontology.git
```
or
```
$ go get github.com/ontio/ontology
```

2) Fetch the dependent third party packages with [Glide](https://glide.sh).

```
$ cd $GOPATH/src/github.com/ontio/ontology
$ glide install
```

3) If necessary, update the dependent third party packages with Glide.

```
$ cd $GOPATH/src/github.com/ontio/ontology
$ glide update
```

4) Build the source code with make.

```
$ make all
```

After building the source code successfully, you should see two executable programs:

- `ontology`: The primary Ontology node application and CLI.
- `tools/sigsvr`: The Ontology Signature Server, `sigsvr` - an RPC server for signing transactions. Detailed documentation can be found [here](https://github.com/ontio/documentation/blob/master/docs/pages/doc_en/Ontology/sigsvr_en.md).

## Run Ontology

You can run Ontology in four different modes:

1) MainNet (`./ontology`)
2) TestNet (`./ontology --networkid 2`)
3) PrivateNet (`./ontology --testmode`)
4) Docker

E.g. for Windows (64-bit), use command prompt and cd to the directory where you installed the Ontology release, then type `start ontology-windows-amd64.exe --networkid 2`. This will sync to TestNet and you can explore further by the help command `ontology-windows-amd64.exe --networkid 2 help`.

### MainNet Sync Node

Run ontology directly

   ```
	./ontology
   ```
then you can connect to Ontology MainNet.

### TestNet Sync Node

Run ontology directly

   ```
	./ontology --networkid 2
   ```

Then you can connect to the Ontology TestNet.

### Local PrivateNet

Create a directory on the host and store the following files in the directory:
- Node program `ontology`
- Wallet file `wallet.dat` (`wallet.dat` can be generated by `./ontology account add -d`)

Run command `$ ./ontology --testmode` can start single-host testnet.

Here's a example of a single-host configuration:

- Directory structure

    ```shell
    $ tree
    └── ontology
        ├── ontology
        └── wallet.dat
    ```

### Run with Docker

Please ensure there is a docker environment in your machine.

1. Make docker image

    - In the root directory of source code, run `make docker`, it will make an Ontology image in docker.

2. Run Ontology image

    - Use command `docker run ontio/ontology` to run Ontology；

    - If you need to allow interactive keyboard input while the image is running, you can use the `docker run -ti ontio/ontology` command to start the image;

    - If you need to keep the data generated by image at runtime, you can refer to the data persistence function of docker (e.g. volume);

    - If you need to add Ontology parameters, you can add them directly after `docker run ontio/ontology` such as `docker run ontio/ontology --networkid 2`.
     The parameters of ontology command line refer to [here](./docs/specifications/cli_user_guide.md).

## Examples

### ONT transfer sample
 -- from: transfer from； -- to: transfer to； -- amount: ONT amount；
```shell
 ./ontology asset transfer  --from=ARVVxBPGySL56CvSSWfjRVVyZYpNZ7zp48 --to=AaCe8nVkMRABnp5YgEjYZ9E5KYCxks2uce --amount=10
```
If the asset transfer is successful, the result will display as follows:

```shell
Transfer ONT
  From:ARVVxBPGySL56CvSSWfjRVVyZYpNZ7zp48
  To:AaCe8nVkMRABnp5YgEjYZ9E5KYCxks2uce
  Amount:10
  TxHash:437bff5dee9a1894ad421d55b8c70a2b7f34c574de0225046531e32faa1f94ce
```
TxHash is the transfer transaction hash, and we can query a transfer result by the TxHash.
Due to block time, the transfer transaction will not be executed before the block is generated and added.

If you want to transfer ONG, just add --asset=ong flag.

Note that ONT is an integer and has no decimals, whereas ONG has 9 decimals. For detailed info please read [Everything you need to know about ONG](https://medium.com/ontologynetwork/everything-you-need-to-know-about-ong-582ed216b870).

```shell
./ontology asset transfer --from=ARVVxBPGySL56CvSSWfjRVVyZYpNZ7zp48 --to=ARVVxBPGySL56CvSSWfjRVVyZYpNZ7zp48 --amount=95.479777254 --asset=ong
```
If transfer of the asset succeeds, the result will display as follows:

```shell
Transfer ONG
  From:ARVVxBPGySL56CvSSWfjRVVyZYpNZ7zp48
  To:AaCe8nVkMRABnp5YgEjYZ9E5KYCxks2uce
  Amount:95.479777254
  TxHash:e4245d83607e6644c360b6007045017b5c5d89d9f0f5a9c3b37801018f789cc3
```

Please note, when you use the address of an account, you can use the index or label of the account instead. Index is the sequence number of a particular account in the wallet. The index starts from 1, and the label is the unique alias of an account in the wallet.

```shell
./ontology asset transfer --from=1 --to=2 --amount=10
```

### Query transfer status sample

```shell
./ontology info status <TxHash>
```

For Example:

```shell
./ontology info status 10dede8b57ce0b272b4d51ab282aaf0988a4005e980d25bd49685005cc76ba7f
```

Result:

```shell
Transaction:transfer success
From:AXkDGfr9thEqWmCKpTtQYaazJRwQzH48eC
To:AYiToLDT2yZuNs3PZieXcdTpyC5VWQmfaN
Amount:10
```

### Query account balance sample

```shell
./ontology asset balance <address|index|label>
```

For Example:

```shell
./ontology asset balance ARVVxBPGySL56CvSSWfjRVVyZYpNZ7zp48
```

or

```shell
./ontology asset balance 1
```
Result:
```shell
BalanceOf:ARVVxBPGySL56CvSSWfjRVVyZYpNZ7zp48
  ONT:989979697
  ONG:28165900
```

For further examples, please refer to the [CLI User Guide](https://ontio.github.io/documentation/cli_user_guide_en.html).

## Contributions

Please open a pull request with a signed commit. We appreciate your help! You can also send your code as email to the developer mailing list. You're welcome to join the Ontology mailing list or developer forum.

Please provide a detailed submission information when you want to contribute code for this project. The format is as follows:

Header line: Explain the commit in one line (use the imperative).

Body of commit message is a few lines of text, explaining things in more detail, possibly giving some background about the issue being fixed, etc.

The body of the commit message can be several paragraphs. Please do proper word-wrap and keep columns shorter than 74 characters or so. That way "git log" will show things  nicely even when it is indented.

Make sure you explain your solution and why you are doing what you are doing, as opposed to describing what you are doing. Reviewers and your future self can read the patch, but might not understand why a particular solution was implemented.

Reported-by: whoever-reported-it +
Signed-off-by: Your Name [youremail@yourhost.com](mailto:youremail@yourhost.com)

## Open source community
### Site

- <https://ont.io/>

### Developer Discord Group

- <https://discord.gg/4TQujHj/>

## License

The Ontology library is licensed under the GNU Lesser General Public License v3.0, read the LICENSE file in the root directory of the project for details.
