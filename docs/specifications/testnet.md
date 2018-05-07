
<h1 align="center">Ontology </h1>
<p align="center" class="version">Version 0.7.0 </p>

[![GoDoc](https://godoc.org/github.com/ontio/ontology?status.svg)](https://godoc.org/github.com/ontio/ontology)
[![Go Report Card](https://goreportcard.com/badge/github.com/ontio/ontology)](https://goreportcard.com/report/github.com/ontio/ontology)
[![Travis](https://travis-ci.org/ontio/ontology.svg?branch=master)](https://travis-ci.org/ontio/ontology)
[![Gitter](https://badges.gitter.im/Join%20Chat.svg)](https://gitter.im/ontio/ontology?utm_source=badge&utm_medium=badge&utm_campaign=pr-badge)

English | [中文](testnet_CN.md) 

## Server deployment
### Select network
To run Ontology successfully,  nodes can be deployed by two ways:

- Public test network(Polaris) sync node deployment
- Single-host deployment
- Multi-hosts deployment

#### Public test network(Polaris) sync node deployment
1.Create account
- Through command line program, create wallet wallet.dat needed for node implementation.
    ```
    $ ./ontology account add -d
    use default value for all options
    Enter a password for encrypting the private key:
    Re-enter password:
    
    Create account successfully.
    Address:  TA9TVuR4Ynn4VotfpExY5SaEy8a99obFPr
    Public key: 120202a1cfbe3a0a04183d6c25ceff1e34957ace6e4899e4361c2e1a2bc3c817f90936
    Signature scheme: SHA256withECDSA
    ```
    Here's a example of host configuration:
   
    Directory structure
    ```shell
    $ tree
    └── ontology
        ├── ontology
        └── wallet.dat
    ```        
2.Start ontology  
  PS: There is no need of config.json file, will use the default setting.

#### Single-host deployment configuration

Create a directory on the host and store the following files in the directory:

- Default configuration file `config.json`
- Node program + Node control program  `ontology`
- Wallet file`wallet.dat`, copy the contents of the configuration file config-solo.config in the root directory to config.json and start the node.
- Edit the config.json file and replace the bookkeeper entries with the public key of your wallet (created above). Use `$ ./ontology account list -v` to get your public key.

Here's a example of single-host configuration:

- Directory structure
    ```shell
    $ tree
    └── ontology
        ├── config.json
        ├── ontology
        └── wallet.dat
    ```

- Set bookkeepers in the config.json file:
    ```
    "Bookkeepers": [ "(public key of your account)1202021c6750d2c5d99813997438cee0740b04a73e42664c444e778e001196eed96c9d" ],
    ```

#### Multi-hosts deployment configuration

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
        ```
        $ ./ontology account add -d
        use default value for all options
        Enter a password for encrypting the private key:
        Re-enter password:
        
        Create account successfully.
        Address:  TA9TVuR4Ynn4VotfpExY5SaEy8a99obFPr
        Public key: 120202a1cfbe3a0a04183d6c25ceff1e34957ace6e4899e4361c2e1a2bc3c817f90936
        Signature scheme: SHA256withECDSA
        ```

5. Bookkeepers configuration

   - While creating a wallet for each node, the public key information of the wallet will be displayed. Fill in the public key information of all nodes in the `Bookkeepers` field of each node's configuration file.

     Note: The public key information for each node's wallet can also be viewed via the command line program:

        ```
        $ ./ontology account list -v
        * 1     TA9TVuR4Ynn4VotfpExY5SaEy8a99obFPr
                Signature algorithm: ECDSA
                Curve: P-256
                Key length: 256 bit
                Public key: 120202a1cfbe3a0a04183d6c25ceff1e34957ace6e4899e4361c2e1a2bc3c817f90936 bit
                Signature scheme: SHA256withECDSA
        ```

        Now multi-host configuration is completed, directory structure of each node is as follows:
        ```
        $ ls
        config.json ontology wallet.dat
        ```

A configuration file fragment can refer to the config-dbft.json file in the root directory.

### Implement

Run each node program in any order and enter the node's wallet password after the `Password:` prompt appears.
```
$ ./ontology
$ - Input your wallet password
```

Run `./ontology --help` for details.

### ONT transfer sample
contract:contract address； - from: transfer from； - to: transfer to； - value: amount；
```shell
  ./ontology asset transfer --caddr=ff00000000000000000000000000000000000001 --value=500 --from  TA6nAAdX77wcsAnuBQxG61zXg3vJUAPpgk  --to TA6Hsjww86b9KBbXFyKEayMcVVafoTGH4K  --password=xxx
```
If transfer asset successd, the result will show as follow:
```
[
	{
		"ContractAddress": "ff00000000000000000000000000000000000001",
		"TxHash": "e0ba3d5807289eac243faceb1a2ac63e8dee4eba208ceac193b0bd606861b729",
		"States": [
			"transfer",
			"TA6nAAdX77wcsAnuBQxG61zXg3vJUAPpgk",
			"TA6Hsjww86b9KBbXFyKEayMcVVafoTGH4K",
			500
		]
	}
]
```