
## Ontology CLI User Guide

- [CLI Account Command](#cli-account-cmd)
- [CLI Info Show Command](#cli-info-show-cmd)
- [CLI Asset Command](#cli-asset-cmd)
- [CLI Setting Command](#cli-setting-cmd)
- [CLI Contract Command](#cli-contract-cmd)

User can start ontology directly,by the command as follow:

```sh
$./ontology
```

The following commands will print the usage of the command or subcommand:

```
./ontology help
./ontology <command> help
```
---

## <a name="cli-account-cmd"></a> CLI Account Command

### Add account

Create a new account contains a key pair and an address.

First select the key type, which could be specified via `-t` or `--type` option.

```sh
$ ontology account add

Select a signature algorithm from the following:

  1  ECDSA
  2  SM2
  3  Ed25519

[default is 1]:
```

If SM2 or EdDSA is selected, the parameters are auto set since each of the two only supports one default setting.

```sh
SM2 is selected.
Use curve sm2p256v1 with key length of 256 bits and SM3withSM2 as the signature scheme.
```
or

```sh
Ed25519 is selected.
Use curve 25519 with key length of 256 bits and Ed25519 as the signature scheme.
```

If ECDSA is selected, the next step is to select the curve:

```sh
Select a curve from the following:

    | NAME  | KEY LENGTH (bits)
 ---|-------|------------------
  1 | P-224 | 224
  2 | P-256 | 256
  3 | P-384 | 384
  4 | P-521 | 521

This determines the length of the private key [default is 2]:
```

The curves determine the key length of 224, 256, 384 and 521 respectively.
This parameter could be specified via `-b` or `--bit-length` option.

Then select a signature scheme:

```sh
Select a signature scheme from the following:

  1  SHA224withECDSA
  2  SHA256withECDSA
  3  SHA384withECDSA
  4  SHA512withECDSA
  5  SHA3-224withECDSA
  6  SHA3-256withECDSA
  7  SHA3-384withECDSA
  8  SHA3-512withECDSA
  9  RIPEMD160withECDSA

This can be changed later [default is 2]:
```

The above selections can be skipped by adding `-d` or `--default` option, which
means using the default parameters.

The private key needs to be encrypted and requires a password:

```sh
Enter a password for encrypting the private key:
Re-enter password:
```

It can be re-encrypted later using the `encrypt` command.

After all the parameters are selected, it will generate a key pair.

The public key will be converted to generate the address. Then output the public informations.

```sh
Create account successfully.
Label: `label you have set`
Address: `base58-address-string`
Public key: `hex-string`
Signature scheme: SHA256withECDSA
```

#### [Other Options]

* **set label**

A label can be set for an account by `-l` or `--label` option,
if this flag wasn't set, new account's label will be blank.
Don't worry if you forget to set label, you can set label with command `account set`.
Look for more informations in [`set`](#account-set) section.

```sh
$ ontology account add -l newaccount
$ ontology account add -l "new account"
```

* **add multiple account**

Multiple accounts can be added by `-n` or `--number` option.
Notice that the number must between 1 to 100, or it will be set to 1 by default.
All the n accounts will use the same params in your account create process,
except the key pair.

```sh
$ ontology account add -n 10
```

### List existing account

```sh
$ ontology account list
* 1  xxxxx yyyyy
  2  xxxxx yyyyy
```

The `*` indicates the default account. The `xxxxx` is address and `yyyyy` is label.

With `-v` option, details of each account would be displayed.

```sh
$ ontology account list -v
* 1 xxxxx
    Label: xxxx
    Signature algorithm: ECDSA
    Curve: P-256
    Key length: 256 bits
    Public key: xxxx
    Signature Scheme: SHA256withECDSA

  ...
```

### <a name="account-set"></a>Modify account

Modify the account settings, such as the signature scheme, or set the account
as default.

Use `-s` or `--signature-scheme` to set the signature scheme for the account.

Use `-d` or `--as-default` to set the account as a default account.

Use `-l` or `--label` to set the label for the account.

Account is specified by the index displayed in the `list` command.

```sh
$ ontology account set -s SHA256withECDSA 1
SHA256withECDSA is selected.

$ ontology account set -d 2
Set account 2 as the default account

$ ontology account set -l "hello world"
Account <1>: label is set to 'hello world'.
```


### Delete account

Delete an existing account by specifying the index.
**Password** must be enter when delete an account.

```sh
$ ontology account del 1
Password:
Delete account successfully.
index = 1, address = xxx, label=xxx
```

### Re-encrypt account

Change the password for an account.

```sh
$ ontology account encrypt 1

Please enter the original password:
```

Then input the new password if the original password is correct.

```sh
Enter a password for encrypting the private key:
Re-enter password:
```

### Import account

Import accounts from some source file to current wallet file.
Use `-s` or `--source` option to specify the source file that your new accounts are from.
Use `-f` or `--file` option to specify your current wallet file where your new accoounts will imported into.
The `-f` can be ignore and default wallet file "wallet.dat" will be used.

```sh
$ ontology account import -s xxx.dat -f wallet.dat

Import finished. 4 accounts has been imported.
```
> **Tips:**
>
> `-f` `--file` can be used for each subcommand of **account**,
> it specifies the wallet file, if not set, default value will be 'wallet.dat'.


## <a name="cli-info-show-cmd"></a>CLI Info Show Command

```
Usage:
    ontology info [command options] [args]

Description:
    With ontology info, you can look up blocks, transactions, etc.

Command:
    version
    block
        --hash value                  block hash value
        --height value                block height value
    tx
        --hash value                  transaction hash value
```

### Example for info version
```
$ ./ontology info version
Result will show as follow:
Node version: 653c-dirty
```

### Example for info block (with height)
```
$ ./ontology info block --height 12
Result will show as follow:
{
	"Action": "getblockbyheight",
	"Desc": "SUCCESS",
	"Error": 0,
	"Result": {
		"Hash": "805e8ad8de2b28884656a393143dc07fea1cc83ddd849e1bbab814c476c26bb2",
		"Header": {
			"BlockRoot": "1ff293c1e15e1a37d449bba5aebda52280ceee645547f9f219ea903f7ad18386",
			"Bookkeepers": [
				"12020352f5ea426d333bd4c0f174fc93e84348513a8d3778ac0c3beaca579cf1937696"
			],
			"ConsensusData": 5922267418785983000,
			"Hash": "805e8ad8de2b28884656a393143dc07fea1cc83ddd849e1bbab814c476c26bb2",
			"Height": 12,
			"NextBookkeeper": "TACATSTXT4E6UJHY57Ccv2TNBPGjbD9hfU",
			"PrevBlockHash": "bdbe264712d38f6d46bffdea8f6a9ab3ef73d343b043d31a0a0594758b73da1d",
			"SigData": [
				"016f3669a1998f2716e91093ca7946263433f92605ae9e6bf2b695145dbe24de5b38c93f8e91ae1d1f89626a4e87ac101f4344fb312c3e482b105a94f7d15f3fe9"
			],
			"Timestamp": 1524280951,
			"TransactionsRoot": "a750c36fca5922a6d4fe22d6b8664faa06f37d4c79c4624a91b10198322adf0e",
			"Version": 0
		},
		"Transactions": [
			{
				"Attributes": [],
				"Fee": [],
				"Hash": "a750c36fca5922a6d4fe22d6b8664faa06f37d4c79c4624a91b10198322adf0e",
				"NetworkFee": 0,
				"Nonce": 0,
				"Payload": {
					"Nonce": 1524280951763615000
				},
				"Sigs": [
					{
						"M": 1,
						"PubKeys": [
							"12020352f5ea426d333bd4c0f174fc93e84348513a8d3778ac0c3beaca579cf1937696"
						],
						"SigData": [
							"01885e3078fe398987c5d89d650714d4e4ab9432ea5bcb98505cf4913d224920203c2f98b1ea24857e6af0f06b3bdf72d862bbe3b0989937cbe06c91775e2aa6e8"
						]
					}
				],
				"TxType": 0,
				"Version": 0
			}
		]
	},
	"Version": "1.0.0"
}

NOTE:
If a invalid value of height is given, the result will show as follow:
{
	"Action": "getblockbyheight",
	"Desc": "UNKNOWN BLOCK",
	"Error": 44003,
	"Result": "",
	"Version": "1.0.0"
}
```

### Example for info block (with hash)
```
$ ./ontology info block --hash 805e8ad8de2b28884656a393143dc07fea1cc83ddd849e1bbab814c476c26bb2
Result will show as follow:
{
	"Action": "getblockbyheight",
	"Desc": "SUCCESS",
	"Error": 0,
	"Result": {
		"Hash": "805e8ad8de2b28884656a393143dc07fea1cc83ddd849e1bbab814c476c26bb2",
		"Header": {
			"BlockRoot": "1ff293c1e15e1a37d449bba5aebda52280ceee645547f9f219ea903f7ad18386",
			"Bookkeepers": [
				"12020352f5ea426d333bd4c0f174fc93e84348513a8d3778ac0c3beaca579cf1937696"
			],
			"ConsensusData": 5922267418785983000,
			"Hash": "805e8ad8de2b28884656a393143dc07fea1cc83ddd849e1bbab814c476c26bb2",
			"Height": 12,
			"NextBookkeeper": "TACATSTXT4E6UJHY57Ccv2TNBPGjbD9hfU",
			"PrevBlockHash": "bdbe264712d38f6d46bffdea8f6a9ab3ef73d343b043d31a0a0594758b73da1d",
			"SigData": [
				"016f3669a1998f2716e91093ca7946263433f92605ae9e6bf2b695145dbe24de5b38c93f8e91ae1d1f89626a4e87ac101f4344fb312c3e482b105a94f7d15f3fe9"
			],
			"Timestamp": 1524280951,
			"TransactionsRoot": "a750c36fca5922a6d4fe22d6b8664faa06f37d4c79c4624a91b10198322adf0e",
			"Version": 0
		},
		"Transactions": [
			{
				"Attributes": [],
				"Fee": [],
				"Hash": "a750c36fca5922a6d4fe22d6b8664faa06f37d4c79c4624a91b10198322adf0e",
				"NetworkFee": 0,
				"Nonce": 0,
				"Payload": {
					"Nonce": 1524280951763615000
				},
				"Sigs": [
					{
						"M": 1,
						"PubKeys": [
							"12020352f5ea426d333bd4c0f174fc93e84348513a8d3778ac0c3beaca579cf1937696"
						],
						"SigData": [
							"01885e3078fe398987c5d89d650714d4e4ab9432ea5bcb98505cf4913d224920203c2f98b1ea24857e6af0f06b3bdf72d862bbe3b0989937cbe06c91775e2aa6e8"
						]
					}
				],
				"TxType": 0,
				"Version": 0
			}
		]
	},
	"Version": "1.0.0"
}

NOTE:
If a invalid value of block hash is given, the result will show as follow:
{
	"Action": "getblockbyhash",
	"Desc": "UNKNOWN BLOCK",
	"Error": 44003,
	"Result": "",
	"Version": "1.0.0"
}
```

### Example for info transaction
```
$ ./ontology info tx --hash 805e8ad8de2b28884656a393143dc07fea1cc83ddd849e1bbab814c476c26bb2
Result will show as follow:
{
	"Action": "gettransaction",
	"Desc": "SUCCESS",
	"Error": 0,
	"Result": {
		"Attributes": [],
		"Fee": [],
		"Hash": "a750c36fca5922a6d4fe22d6b8664faa06f37d4c79c4624a91b10198322adf0e",
		"NetworkFee": 0,
		"Nonce": 0,
		"Payload": {
			"Nonce": 1524280951763615000
		},
		"Sigs": [
			{
				"M": 1,
				"PubKeys": [
					"12020352f5ea426d333bd4c0f174fc93e84348513a8d3778ac0c3beaca579cf1937696"
				],
				"SigData": [
					"01885e3078fe398987c5d89d650714d4e4ab9432ea5bcb98505cf4913d224920203c2f98b1ea24857e6af0f06b3bdf72d862bbe3b0989937cbe06c91775e2aa6e8"
				]
			}
		],
		"TxType": 0,
		"Version": 0
	},
	"Version": "1.0.0"
}

NOTE:
If a invalid value of transaction hash is given, the result will show as follow:
{
	"Action": "gettransaction",
	"Desc": "UNKNOWN TRANSACTION",
	"Error": 44001,
	"Result": "",
	"Version": "1.0.0"
}
```
---

## <a name="cli-asset-cmd"></a> CLI Asset Command
```
Usage:
    ontology asset [command options] [args]

Description:
    With this command, you can control assert through transaction.

Command:
    transfer
        --caddr     value                 smart contract address
        --from      value                 wallet address base58, which will transfer from
        --to        value                 wallet address base58, which will transfer to
        --value     value                 how much asset will be transfered
        --password  value                 use password who transfer from
    status
        --hash     value                  transfer transaction hash
```

### Example for asset transfer
```
$ ./ontology asset transfer --caddr=ff00000000000000000000000000000000000001 --value=500 --from TA6VvtGekMfinP97CTL9SH5WTowChUungL   --to TA7bRMkjMbP4JHhAVASN1fwpB56bi8aYUK  --password passwordtest
If transfer asset successd, the result will show as follow:
[
	{
		"ContractAddress": "ff00000000000000000000000000000000000001",
		"TxHash": "e0ba3d5807289eac243faceb1a2ac63e8dee4eba208ceac193b0bd606861b729",
		"States": [
			"transfer",
			"TA6VvtGekMfinP97CTL9SH5WTowChUungL",
			"TA7bRMkjMbP4JHhAVASN1fwpB56bi8aYUK",
			500
		]
	}
]

Otherwise, the null value will given.
```

### Example for asset status
```
$ ./ontology asset status --hash e0ba3d5807289eac243faceb1a2ac63e8dee4eba208ceac193b0bd606861b729
If query succeed, the result will show as follow:
[
	{
		"ContractAddress": "ff00000000000000000000000000000000000001",
		"TxHash": "e0ba3d5807289eac243faceb1a2ac63e8dee4eba208ceac193b0bd606861b729",
		"States": [
			"transfer",
			"TA6VvtGekMfinP97CTL9SH5WTowChUungL",
			"TA7bRMkjMbP4JHhAVASN1fwpB56bi8aYUK",
			500
		]
	}
]

Otherwise, the null value will given.
```

---

## <a name="cli-setting-cmd"></a> CLI Setting Command
```
Usage:
    ontology set [command options] [args]

Description:
    With ontology set, you can configure the node.

Command:
    --debuglevel value                 debug level(0~6) will be set
    --consensus value                  [ on / off ]
```

### Example for setting debuglevel
```
$ ./ontology set --debuglevel 1
When setting succeed, the result will show as follow:
map[desc:SUCCESS error:0 id:0 jsonpc:2.0 result:true]
```

### Example for setting consensus
```
$ ./ontology set --consensus on
When setting succeed, the result will show as follow:
map[desc:SUCCESS error:0 id:0 jsonpc:2.0 result:true]
```
---


## <a name="cli-contract-cmd"></a> CLI Contract Command
```
Usage:
    ontology contract [command options] [args]

Description:
    With this command, you can invoke a smart contract

Command:
    invoke
        --caddr      value               smart contract address that will be invoke
        --params     value               params will be

    deploy
        --type       value               contract type ,value: 1 (NEOVM) | 2 (WASM)
        --store      value               does this contract will be stored, value: true or false
        --code       value               directory of smart contract that will be deployed
        --cname      value               contract name that will be deployed
        --cversion   value               contract version which will be deployed
        --author     value               owner of deployed smart contract
        --email      value               owner email who deploy the smart contract
        --desc       value               contract description when deploy one
```

### Example for contract deploy
```
$ ./ontology contract deploy --type=1 --store=true --code=./deploy.txt --cname=testContract --cversion=0.0.1 --author=Yihen --desc="this is my first test contract" --email="name@emailaddr.com"
User need to input password, when deployed succeed, the result will show as follow:
Deploy smartContract transaction hash: b68307e7659a53a10d3061a3e4721b5a152bd4e6e80b75002f5cc8b5294be493
```

### Example for contract invoke
```
$ ./ontology contract invoke --caddr TA6VvtGekMfinP97CTL9SH5WTowChUungL --params contract-params-need
User need to input password, when invoke succeed, the result will show as follow:
invoke transaction hash:64976014daec682f8e3a4fc62c26f4636ac7be6bd548f67f24f62211394f8923
```
