# Ontology CLI Instruction

[English|[中文](cli_user_guide_CN.md)]

Ontology CLI is an Ontology node command line client for starting and managing Ontology nodes, managing user wallets, sending transactions, and deploying and invoking contracts.

* [Ontology CLI Instruction](#ontology-cli-instruction)
	* [1. Start and Manage Ontology Nodes](#1-start-and-manage-ontology-nodes)
		* [1.1 Startup Parameters](#11-startup-parameters)
			* [1.1.1 Ontology System Parameters](#111-ontology-system-parameters)
			* [1.1.2 Account Parameters](#112-account-parameters)
			* [1.1.3 Consensus Parameters](#113-consensus-parameters)
			* [1.1.4 P2P Network Parameters](#114-p2p-network-parameters)
			* [1.1.5 RPC Server Parameters](#115-rpc-server-parameters)
			* [1.1.6 RESTful Server Parameters](#116-restful-server-parameters)
			* [1.1.7 Web Socket Server Parameters](#117-web-socket-server-parameters)
			* [1.1.8 Test Mode Parameters](#118-test-mode-parameters)
			* [1.1.9 Transaction Parameters](#119-transaction-parameter)
		* [1.2 Node Deployment](#12-node-deployment)
			* [1.2.1 MainNet Bookkeeping Node Deployment](#121-mainnet-bookkeeping-node-deployment)
			* [1.2.2 MainNet Synchronization Node Deployment](#122-mainnet-synchronization-node-deployment)
			* [1.2.3 Deploying on public test network Polaris sync node](#123-deploying-on-public-test-network-polaris-sync-node)
			* [1.2.4 Single-Node Test Network Deployment](#124-single-node-test-network-deployment)
	* [2. Wallet Management](#2-wallet-management)
		* [2.1. Add Account](#21-add-account)
			* [2.1.1 Add Account Parameters](#211-add-account-parameters)
		* [2.2 View Account](#22-view-account)
		* [2.3 Modify Account](#23-modify-account)
			* [2.3.1 Modifying Account Parameters](#231-modifying-account-parameters)
		* [2.4 Delete Account](#24-delete-account)
		* [2.5 Import Account](#25-import-account)
			* [2.5.1 Import Account Parameters](#251-import-account-parameters)
			* [2.5.2 Import Account by WIF](#252-import-account-by-wif)
	* [3. Asset Management](#3-asset-management)
		* [3.1 Check Your Account Balance](#31-check-your-account-balance)
		* [3.2 ONT/ONG Transfers](#32-ontong-transfers)
			* [3.2.1 Transfer Arguments](#321-transfer-arguments)
		* [3.3 Authorize Transfer](#33-authorize-transfer)
			* [3.3.1 Authorize Transfer Parameter](#331-authorize-transfer-parameter)
		* [3.4 View Authorized Transfer Balance](#34-view-authorized-transfer-balance)
			* [3.4.1 View Authorized Transfer Balance Parameters](#341-view-authorized-transfer-balance-parameters)
		* [3.5 Transferring Fund from Authorized Accounts](#35-transferring-fund-from-authorized-accounts)
			* [3.5.1 Transferring Funds from Authorized Accounts Parameters](#351-transferring-fund-from-authorized-accounts-parameters)
		* [3.6 View Unlocked ONG Balance](#36-view-unlocked-ong-balance)
		* [3.7 Extract Unlocked ONG](#37-extract-unlocked-ong)
			* [3.7.1 Extracting Unlocked ONG Parameters](#371-extracting-unlocked-ong-parameters)
	* [4 Query Information](#4-query-information)
		* [4.1 Query Block Information](#41-query-block-information)
		* [4.2 Query Transaction Information](#42-query-transaction-information)
		* [4.3 Query Transaction Execution Information](#43-query-transaction-execution-information)
	* [5. Smart Contract](#5-smart-contract)
		* [5.1 Smart Contract Deployment](#51-smart-contract-deployment)
			* [5.1.1 Smart Contract Deployment Parameters](#511-smart-contract-deployment-parameters)
		* [5.2 Smart Contract Execution](#52-smart-contract-execution)
			* [5.2.1 Smart Contract Execution Parameters](#521-smart-contract-execution-parameters)
		* [5.3 Smart Contract Code Execution Directly](#53-smart-contract-code-execution-directly)
			* [5.3.1 Smart Contract Code Execution Directly Parameters](#531-smart-contract-code-execution-directly-parameters)
	* [6. Block Import and Export](#6-block-import-and-export)
		* [6.1 Export Blocks](#61-export-blocks)
			* [6.1.1 Export Block Parameters](#611-export-block-parameters)
		* [6.2 Import Blocks](#62-import-blocks)
			* [6.2.1 Importing Block Parameters](#621-importing-block-parameters)
	* [7、Build Transaction](#7-build-transaction)
		* [7.1 Build Transfer Transaction](#71-build-transfer-transaction)
			* [7.1.1 Build Transfer Transaction Parameters](#711-build-transfer-transaction-params)
		* [7.2 Build Authorize Transfer](#72-build-authorize-transfer)
			* [7.2.1 Build Authorize Transfer Params](#721-build-authorize-transfer-params)
		* [7.3 Build Transfer From Authorize Account Transaction](#73-build-transfer-from-authorize-account-transaction)
			* [7.3.1 Build Transfer From Authorize Account Transaction Params](#731-build-transfer-from-authorize-account-transaction-params)
		* [7.4 Build Withdraw ONG Transaction](#74-build-withdraw-ong-transaction)
			* [7.4.1 Build Withdraw ONG Transaction Params](#741-build-withdraw-ong-transaction-params)
	* [8. Sign To Transaction](#8-sign-to-transaction)
		* [8.1 Sign To Transaction Parameters](#81-sign-to-transaction-parameters)
	* [9. Generate Multi-Signature Address](#9-generate-multi-signature-address)
		* [9.1 Generate Multi-Signature Address Parameters](#91-generate-multi-signature-address-parameters)
	* [10. Multi-Signature To Transaction](#10-multi-signature-to-transaction)
		* [10.1 Multi-signature parameters for transactions](#101-对交易多重签名参数)
	* [11. Send Transaction](#11-send-transaction)
		* [11.1 Send Transaction Parameters](#111-send-transaction-parameters)
	* [12. Show Transaction Infomation](#12-show-transaction-infomation)

## 1. Start and Manage Ontology Nodes

Ontology CLI has a lot of startup parameters for configuring some of the Ontology node's behavior. Use ./Ontology -help to see all startup parameters supported by the Ontology CLI node. If Ontology CLI is started without any parameters, it will access the Ontology main network as a synchronous node by default.

### 1.1 Startup Parameters

The following are the command line parameters supported by Ontology CLI:

#### 1.1.1 Ontology System Parameters

--config
The config parameter specifies the file path of the genesis block for the current Ontolgy node. If not specified, Ontology will use the config of Polaris TestNet. Note that the genesis block configuration must be the same for all nodes in the same network, otherwise it will not be able to synchronize blocks or start nodes due to block data incompatibility.

--loglevel
The loglevel parameter is used to set the log level the Ontology outputs. Ontology supports 7 different log levels, i.e. 0:Trace 1:Debug 2:Info 3:Warn 4:Error 5:Fatal 6:MaxLevel. The logs are logged from low to high, and the log output volume is from high to low. The default value is 2, which means that only logs at the info level or higher level.

--disable-event-log
The disable-event-log parameter is used to disable the event log output when the smart contract is executed to improve the node transaction execution performance. The Ontology node enables the event log output function by default.

--data-dir
The data-dir parameter specifies the storage path of the block data. The default value is "./Chain".

#### 1.1.2 Account Parameters

--wallet, -w
The wallet parameter is used to specify the wallet file path when the Ontology node starts. The default value is "./wallet.dat".

--account, -a
The account parameter is used to specify the account address when the Ontology node starts. If the account is null, it uses the wallet default account.

--password, -p
The password parameter is used to specify the account password when Ontology node starts. Because the account password entered in the command line is saved in the log, it is easy to leak the password. Therefore, it is not recommended to use this parameter in a production environment.

#### 1.1.3 Consensus Parameters

--enable-consensus
The enable-consensus parameter is used to turn the consensus on. If the current node will startup as a bookkeeper node, this flag must be enabled. The default is disable.

--max-tx-in-block
The max-tx-in-block parameter is used to set the maximum transaction number of a block. The default value is 50000.

#### 1.1.4 P2P Network Parameters

--networkid
The networkid parameter is used to specify the network ID. Different networkids cannot connect to the blockchain network. 1=main net, 2=polaris test net, 3=testmode, and other for custom network.

--nodeport
The nodeport parameter is used to specify the P2P network port number. The default value is 20338.

--consensusport
The consensusport parameter specifies the consensus network port number. By default, the consensus network reuses the P2P network, so it is not necessary to specify a consensus network port. After the dual network is enabled with the --dual-port parameter, the consensus network port number must be set separately. The default is 20339.

--dual-port
The dual-port parameter initiates a dual network, i.e. a P2P network for processing transaction messages and a consensus network for consensus messages. The parameter disables by default.

--httpinfo-port
httpinfo-port parameter specifies the http server port of viewing node information. The default value is 0 which means closes the http server.

#### 1.1.5 RPC Server Parameters

--disable-rpc
The disable-rpc parameter is used to shut down the EPC server. The Ontology node starts the RPC server by default at startup.

--rpcport
The rpcport parameter specifies the port number to which the RPC server is bound. The default is 20336.

#### 1.1.6 RESTful Server Parameters

--rest
The rest parameter is used to start the RESTful server.

--restport
The restport parameter specifies the port number to which the RESTful server is bound. The default value is 20334.

--restmaxconns
The restconnlimit parameter specifies the maximum connections. The default value is 1024.

#### 1.1.7 WebSocket Server Parameters

--ws
The ws parameter is used to start the WebSocket server.

--wsport
The wsport parameter specifies the port number to which the WebSocket server is bound. The default value is 20335

#### 1.1.8 Test Mode Parameters

--testmode
The testmode parameter is used to start a single node test network for ease of development and debug. In testmode, Ontology will start RPC, RESTful and WebSocket server, and blockchain data will be clear generated by the last start in testmode.

--testmode-gen-block-time
The testmode-gen-block-time parameter is used to set the block-out time in test mode. The time unit is in seconds, and the minimum block-out time is 2 seconds.

#### 1.1.9 Transaction Parameter

--gasprice
The gasprice parameter is used to set the lowest gasprice of the current node transaction pool to accept transactions. Transactions below this gasprice will be discarded. The default value is 500(0 in testmode).

--gaslimit
The gaslimit parameter is used to set the gaslimit of the current node transaction pool to accept transactions. Transactions below this gaslimit will be discarded. The default value is 20000.

--disable-tx-pool-pre-exec
The disable-tx-pool-pre-exec parameter is used to disable preExecution of a transaction from network in the transaction pool. By default, preExecution is enabled when ontology bootstrap.

--disable-sync-verify-tx
The disable-sync-verify-tx is used to disable sync verify transaction in send transaction,include rpc restful websocket.

--disable-broadcast-net-tx
The disable-broadcast-net-tx is used to disable broadcast a transaction from network in the transaction pool. By default, this function is enabled when ontology bootstrap.

### 1.2 Node Deployment

#### 1.2.1 MainNet Bookkeeping Node Deployment

According to different roles of nodes, they can be divided into bookkeeping nodes and synchronization nodes. Bookkeeping nodes participate in the network consensus, and synchronization nodes only synchronize the blocks generated by the bookkeeping nodes. Since Ontology node won't start consensus by default, consensus must be turned on by the --enable-consensus parameter. The Ontology node will start the RPC server by default and output the event log of the smart contract. Therefore, if there is no special requirement, you can use the --disable-rpc and --disable-event-log command line parameters to turn off the RPC and eventlog modules.

Recommended bookkeeping node startup parameters:

```
./Ontology --enbale-consensus --disable-rpc --disable-event-log
```
    - `enbale-consensus` is use to start the consensus
    - `disable-rpc` is to close the rpc services for the safe concerns.
    - `disable-event-log` is to disable the event log for high performance.
If the node does not use the default genesis block configuration file and wallet account, the node can specify them with the --config, --wallet, --account parameters.
At the same time, if the bookkeeping node needs to modify the default minimum gas price and gas limit of the transaction pool, it can set the parameters by --gasprice and --gaslimit.

#### 1.2.2 MainNet Synchronization Node Deployment

Since the synchronization node only synchronizes the blocks generated by the bookkeeping node and does not participate in the network consensus.

```
./Ontology
```

If the node does not use the default genesis block configuration file, it can be specified with the --config parameter. Since consensus won't be turned on, you don't need a wallet when starting up a synchronization node.

#### 1.2.3 Deploying on public test network Polaris sync node

Run ontology straightly

```
./Ontology --networkid 2
```
#### 1.2.4 Single-Node Test Network Deployment

Ontology supports single-node network deployment for the development of test environments. To start a single-node test network, you only need to add the --testmode command line parameter.

```
./Ontology --testmode
```

If the node does not use the default genesis block configuration file and wallet account, the node can specify them with the --config, --wallet, --account parameters.
At the same time, if the bookkeeping node needs to modify the default minimum gas price and gas limit of the transaction pool, it can set the parameters by --gasprice and --gaslimit.

Note that, Ontology will turn consensus RPC, RESTful, and WebSocket server on in test mode.

## 2. Wallet Management

Wallet management commands can be used to add, view, modify, delete, and import account.
You can use ./Ontology account --help command to view help information of wallet management command.

### 2.1. Add Account

Ontology supports multiple encryption algorithms, including ECDSA, SM2, and ED25519.

When using ECDSA encryption algorithm, it can support multiple key curves, such as: P-224, P-256, P-384, P-521; In addition, when using ECDSA encryption algorithm, you can also specify the signature scheme such as: SHA224withECDSA, SHA256withECDSA, SHA384withECDSA, SHA512withEdDSA, SHA3-224withECDSA, SHA3-256withECDSA, SHA3-384withECDSA, SHA3-512withECDSA, RIPEMD160withECDSA.

When using the SM2 encryption algorithm, the sm2p256v1 curve and SM3withSM2 signature algorithm will be used.

When using the ED25519 encryption algorithm, the 25519 curve and SHA512withEdDSA signature algorithm will be used.


**Default account**

Each wallet has a default account, which is generally the first account added. The default account cannot be deleted, you can modify the default account by ./Ontology account set command.


#### 2.1.1 Add Account Parameters

--type,t
The type parameter is used to set the encryption algorithm and supports the ecdsa, sm2, and ed25519 encryption algorithms.

--bit-length,b
bit-length parameter is used to specify the key length. If ecdsa is the encryption algorithm, you can choose p-224, p-256, p-384 or p-521; if sm2 is the encryption algorithm, the default is sm2p256v1; if ed25519 is the encryption algorithm, the default is 25519.

--signature-scheme,s
The signature-scheme parameter is used to specify the key signature scheme. For the ecdsa encryption algorithm, these signature schemes such as SHA224withECDSA, SHA256withECDSA, SHA384withECDSA, SHA512withEdDSA, SHA3-224withECDSA, SHA3-256withECDSA, SHA3-384withECDSA, SHA3-512withECDSA, RIPEMD160withECDSA are supported; For the sm2 encryption algorithm, SM3withSM2 signature scheme is used by default. If ed25519 is the encryption algorithm, the SHA512withEdDSA signature scheme is used by default.

--default
The default parameter uses the system's default key scheme. The default key scheme will use the ECDSA encryption algorithm with P-256 curve and SHA256withECDSA as the signature algorithm.

--label
Label is used to set labels for newly created accounts for easy and fast use of accounts. Note that duplicate label names cannot appear in the same wallet file. An account with no label is an empty string.

--wallet
The wallet parameter specifies the wallet file path. If the wallet file does not exist, a new wallet file will be automatically created.

--number
The number parameter specifies the number of accounts that need to be created. You can batch create accounts by number parameter. The default value is 1.

--ontid
The parameter is used to create ONT ID instead of account.

**Add account**

```
./Ontology account add --default
```

You can view the help information by ./Ontology account add --help.

### 2.2 View Account

Command：

```
./Ontology account list
```

You can view all account information in your current wallet. such as:

```
$ ./Ontology account list
Index:1    Address:TA587BCw7HFwuUuzY1wg2HXCN7cHBPaXSe  Label: (default)
Index:2    Address:TA5gYXCSiUq9ejGCa54M3yoj9kfMv3ir4j  Label:
```
Among them, Index is the index of the account in the wallet and the index starts from 1. Address is the address of the account. Label is the label of the account, default indicates that the current account is the default account.
In Ontology CLI, you can find accounts by Index, Address, or a non-empty Label.

Use --v to view the details of the account. You can view the help information via ./Ontology account list --help.

### 2.3 Modify Account

Using the modify account command to modify the account's label, reset the default account, modify the account's password. If the account is the key of the ECDSA encryption algorithm, you can also modify the key's signature scheme. You can view the help information via ./Ontology account add --help.


#### 2.3.1 Modifying Account Parameters

--as-default, -d
The as-default parameter sets the account as the default account. A wallet only has one default account. After setting a new default account, the previous default account will automatically cancel the default account properties.


--wallet, -w
The wallet parameter specifies the wallet path. The default value is "./wallet.dat".

--label, -l
The label parameter is used to set a new label for the account. Note that a wallet file cannot have the same label.

--change-passwd
The change-passwd parameter is used to modify the account password.

--signature-scheme, -s
The signature-scheme parameter is used to modify the account signature scheme. If the account uses an ECDSA key, the following ECDSA-supported signature schemes can be modified: SHA224withECDSA, SHA256withECDSA, SHA384withECDSA, SHA512withEdDSA, SHA3-224withECDSA, SHA3-256withECDSA, SHA3-384withECDSA, SHA3-512withECDSA, RIPEMD160withECDSA.

**Set default account**

```
./Ontology account set --d <address|index|label>
```
**Edit account label**

```
./Ontology account set --label=XXX <address|index|label>
```
**Change account password**

```
./Ontology account set --changepasswd <address|index|label>
```

**Modify ECDSA key signature scheme**

```
./Ontology account set --s=SHA256withECDSA <address|index|label>
```
### 2.4 Delete Account

Unnecessary accounts in the wallet can be deleted and cannot be recovered after delete. Note: The default account cannot be deleted.

```
/Ontology account del <address|index|label>
```
### 2.5 Import Account

The import account command can import account of another wallet into the current wallet.

#### 2.5.1 Import Account Parameters

--wallet,w
The wallet parameter specifies the current wallet path for saving the wallet-introduced account.

--source,s
The source parameter specifies the imported wallet path.

```
./Ontology account import -s=./source_wallet.dat
```

#### 2.5.2 Import Account by WIF
Fill the WIF into a text file, and use the cmd below to import the key
ontology account import --wif --source key.txt

## 3. Asset Management

Asset management commands can check account balance, ONT/ONG transfers, extract ONG, and view unbound ONG.

### 3.1 Check Your Account Balance

```
./Ontology asset balance <address|index|label>
```
### 3.2 ONT/ONG Transfers

#### 3.2.1 Transfer Arguments

--wallet, -w
Wallet specifies the transfer-out account wallet path. The default value is: "./wallet.dat".

--gasprice
The gasprice parameter specifies the gas price of the transfer transaction. The gas price of the transaction cannot be less than the lowest gas price set by node's transaction pool, otherwise the transaction will be rejected. The default value is 500 (0 in testmode). When there are transactions that are queued for packing into the block in the transaction pool, the transaction pool will deal with transactions according to the gas price and transactions with high gas prices will be prioritized.

--gaslimit
The gaslimit parameter specifies the gas limit of the transfer transaction. The gas limit of the transaction cannot be less than the minimum gas limit set by the node's transaction pool, otherwise the transaction will be rejected. Gasprice * gaslimit is actual ONG costs. The default value is 20000.

--asset
The asset parameter specifies the asset type of the transfer. ont indicates the ONT and ong indicates the ONG. The default value is ont.

--from
The from parameter specifies the transfer-out account address.

--to
The to parameter specifies the transfer-in account address.

--amount
The amount parameter specifies the transfer amount. Note: Since the precision of the ONT is 1, if the input is a floating-point value, then the value of the fractional part will be discarded; the precision of the ONG is 9, so the fractional part beyond 9 bits will be discarded.

**Transfer**

```
./Ontology asset transfer --from=<address|index|label> --to=<address|index|label> --amount=XXX --asset=ont
```

### 3.3 Authorize Transfer

A user may authorize others to transfer money from his account, and he can specify the transfer amount when authorizing the transfer.

#### 3.3.1 Authorize Transfer Parameter

--wallet, -w
Wallet specifies the transfer-out account wallet path. The default value is: "./wallet.dat".

--gasprice
The gasprice parameter specifies the gas price of the transfer transaction. The gas price of the transaction cannot be less than the lowest gas price set by node's transaction pool, otherwise the transaction will be rejected. The default value is 500 (0 in testmode). When there are transactions that are queued for packing into the block in the transaction pool, the transaction pool will deal with transactions according to the gas price and transactions with high gas prices will be prioritized.

--gaslimit
The gaslimit parameter specifies the gas limit of the transfer transaction. The gas limit of the transaction cannot be less than the minimum gas limit set by the node's transaction pool, otherwise the transaction will be rejected. Gasprice * gaslimit is actual ONG costs. The default value is 20000.

--asset
The asset parameter specifies the asset type of the transfer. ont indicates the ONT and ong indicates the ONG. The default value is ont.

--from
The from parameter specifies the transfer-out account address.

--to
The to parameter specifies the transfer-in account address.

--amount
The amount parameter specifies the transfer amount. Note: Since the precision of the ONT is 1, if the input is a floating-point value, then the value of the fractional part will be discarded; the precision of the ONG is 9, so the fractional part beyond 9 bits will be discarded.


**Authorize Transfer**

```
./Ontology asset approve --from=<address|index|label> --to=<address|index|label> --amount=XXX --asset=ont
```

### 3.4 View Authorized Transfer Balance

After authorizing a user to transfer funds, the user can execute transfer operation within the authorized amount multiple times based on needs. The command of checking authorized transfer balances can see the untransferred balances.

#### 3.4.1 View Authorized Transfer Balance Parameters

--wallet, -w
Wallet specifies the transfer-out account wallet path. The default value is: "./wallet.dat".

--asset
The asset parameter specifies the asset type of the transfer. ont indicates the ONT and ong indicates the ONG. The default value is ont.

--from
The from parameter specifies the transfer-out account address.

--to
The to parameter specifies the transfer-in account address.

**View Authorized Transfer Balance**

```
./Ontology asset allowance --from=<address|index|label> --to=<address|index|label>
```

### 3.5 Transferring Funds from Authorized Accounts

After user authorization, the transfer can be made from the authorized account.

#### 3.5.1 Transferring Funds from Authorized Accounts Parameters
--wallet, -w
Wallet specifies the wallet path of authorized account. The default value is: "./wallet.dat".

--gasprice
The gasprice parameter specifies the gas price of the transfer transaction. The gas price of the transaction cannot be less than the lowest gas price set by node's transaction pool, otherwise the transaction will be rejected. The default value is 500 (0 in testmode). When there are transactions that are queued for packing into the block in the transaction pool, the transaction pool will deal with transactions according to the gas price and transactions with high gas prices will be prioritized.

--gaslimit
The gaslimit parameter specifies the gas limit of the transfer transaction. The gas limit of the transaction cannot be less than the minimum gas limit set by the node's transaction pool, otherwise the transaction will be rejected. Gasprice * gaslimit is actual ONG costs. The default value is 20000.

--asset
The asset parameter specifies the asset type of the transfer. ont indicates the ONT and ong indicates the ONG. The default value is ont.

--from
The from parameter specifies the transfer-out account address.

--to
The to parameter specifies the transfer-in account address.

--sender
The sender parameter specifies the account address that actually operates the authorized transfer. If no sender parameter is specified, the sender parameter defaults to the value of the to parameter.

--amount
The amount parameter specifies the transfer amount and the transfer amount cannot be greater than the authorized transfer balance. Otherwise, the transaction will be rejected. Note: Since the precision of the ONT is 1, if the input is a floating-point value, then the value of the fractional part will be discarded; the precision of the ONG is 9, so the fractional part beyond 9 bits will be discarded.

**Transfer from authorized account**

```
./Ontology asset transferfrom --from=<address|index|label> --to=<address|index|label> --sender=<address|index|label> --amount=XXX
```

### 3.6 View Unbound ONG Balance

The ONG adopts the periodical unbinding policy to release the ONG bound to ONT. Use the following command to view the current account unbound ONG balance.

```
./Ontology asset unboundong <address|index|label>
```

### 3.7 Extract Unbound ONG

Use the following command to extract all unbound ONG.

#### 3.7.1 Extracting Unbound ONG Parameters

--wallet, -w
Wallet specifies the wallet path of extracted account. The default value is: "./wallet.dat".

--gasprice
The gasprice parameter specifies the gas price of the transfer transaction. The gas price of the transaction cannot be less than the lowest gas price set by node's transaction pool, otherwise the transaction will be rejected. The default value is 500 (0 in testmode). When there are transactions that are queued for packing into the block in the transaction pool, the transaction pool will deal with transactions according to the gas price and transactions with high gas prices will be prioritized.

--gaslimit
The gaslimit parameter specifies the gas limit of the transfer transaction. The gas limit of the transaction cannot be less than the minimum gas limit set by the node's transaction pool, otherwise the transaction will be rejected. Gasprice * gaslimit is actual ONG costs. The default value is 20000.

**Extract Unbound ONG**
```
./Ontology asset withdrawong <address|index|label>
```
## 4. Query Information

Query information command can query information such as blocks, transactions, and transaction executions. You can use the ./Ontology info block --help command to view help information.

### 4.1 Query Block Information

```
./Ontology info block <height|blockHash>
```

Block information can be queried by block height or block hash.

### 4.2 Query Transaction Information

```
./Ontology info tx <TxHash>
```

You can query transaction information by transaction hash.

### 4.3 Query Transaction Execution Information

```
./Ontology info status <TxHash>
```
You can query the transaction execution information through the transaction hash, and the following example is as follows:

```
{
   "TxHash": "4c00674d96b1d3d2c8152b905cae6f87fff0ec8acf28ca3e7465aac59de814a1",
   "State": 1,
   "GasConsumed": 0,
   "Notify": [
      {
         "ContractAddress": "ff00000000000000000000000000000000000001",
         "States": [
            "transfer",
            "TA587BCw7HFwuUuzY1wg2HXCN7cHBPaXSe",
            "TA5gYXCSiUq9ejGCa54M3yoj9kfMv3ir4j",
            10
         ]
      }
   ]
}
```
Among them, State represents the execution result of the transaction. The value of State is 1, indicating that the transaction execution is successful. When the State value is 0, it indicates that the execution failed. GasConsumed indicates the ONG consumed by the transaction execution. Notify represents the Event log output when the transaction is executed. Different transactions may output different event logs.

## 5. Smart Contract

Smart contract operations support the deployment of NeoVM smart contract, and the pre-execution and execution of NeoVM smart contract.

### 5.1 Smart Contract Deployment

Before smart deployment, you need to compile the NeoVM contract compiler such as [SmartX](http://smartx.ont.io) and save the compiled code in a local text file.

#### 5.1.1 Smart Contract Deployment Parameters

--wallet, -w
The wallet parameter specifies the wallet path of account for deploying smart contracts. Default: "./wallet.dat".

--account, -a
The account parameter specifies the account that a contract deploys.

--gasprice
The gasprice parameter specifies the gas price of the transfer transaction. The gas price of the transaction cannot be less than the lowest gas price set by node's transaction pool, otherwise the transaction will be rejected. The default value is 500 (0 in testmode). When there are transactions that are queued for packing into the block in the transaction pool, the transaction pool will deal with transactions according to the gas price and transactions with high gas prices will be prioritized.

--gaslimit
The gaslimit parameter specifies the gas limit of the transfer transaction. The gas limit of the transaction cannot be less than the minimum gas limit set by the node's transaction pool, otherwise the transaction will be rejected. Gasprice * gaslimit is actual ONG costs. The default value is 20000.

**For contract deployments, the gaslimit value must be greater than 20000000, and there must be sufficient ONG balance in the account.**

--needstore
The needstore parameter specifies whether the smart contract needs to use persistent storage. If needed, this parameter is required. The default is not used.

--code
The code parameter specifies the code path of a smart contract.

--name
The name parameter specifies the name of a smart contract.

--version
The version parameter specifies the version number of a smart contract.

--author
The author parameter specifies the author information of a smart contract.


--email
The email parameter specifies the contact email of a smart contract.

--desc
The desc parameter specifies the description of a smart contract.

--prepare, -p
The prepare parameter indicates that the current deploy is a pre-deploy contract. The transactions executed will not be packaged into blocks, nor will they consume any ONG. Via pre-deploy contract, user can known the the gas limit required for the current deploy.

**Smart Contract Deployment**

```
./Ontology contract deploy --name=xxx --code=xxx --author=xxx --desc=xxx --email=xxx --needstore --gaslimit=100000000
```

After deployment, the TxHash of the transaction and the contract address will be returned. For example:


```
Deploy contract:
  Contract Address:806fbee1fcfb554af47844edd4d4ce2918737747
  TxHash:99d719f51837acfa48f9cd2a21983fb993bc8d5a763b497802f7b872be2338fe
```

You can query the contract execution status with the ./Ontology info status <TxHash> command. If an error such as UNKNOWN TRANSACTION is returned, it means that the transaction has not been posted. The transaction may be queued in the transaction pool to be packaged, or the transaction may be rejected because the gas limit or gas price is set too low.

If the returned execution state - State is equal to 0, it indicates that the transaction execution fails. If State is equal to 1, the transaction execution is successful and the contract deployment is successful. Such as:

```
Transaction states:
{
   "TxHash": "99d719f51837acfa48f9cd2a21983fb993bc8d5a763b497802f7b872be2338fe",
   "State": 1,
   "GasConsumed": 0,
   "Notify": []
}
```

### 5.2 Smart Contract Execution

The NeoVM smart contract supports array, bytearray, string, int, and bool parameter types. Array represents an array of objects, which can nest any number and any type of parameters that NeoVM supports; bytearray represents a byte array, and the input needs to be hexadecimal encoded into a string, such as []byte("HelloWorld"). : 48656c6c6f576f726c64; string represents a string literal; int represents an integer, because the NeoVM virtual machine does not support floating-point values, so it is necessary to convert the floating-point number into an integer; bool represents a Boolean variable, with true, false.

In Ontology CLI, prefix method is used to construct the input parameters. The type of the parameter will be declared before the parameter, such as string input parameters represented as string: hello; integer parameters as int: 10; Boolean parameters represented as bool: true and so on. Multiple parameters are separated by ",". The object numerical array type uses "[ ]" to indicate the array element range, such as [int:10,string:hello,bool:true].

Input parameters example：

```
string:method,[string:arg1,int:arg2]
```

#### 5.2.1 Smart Contract Execution Parameters

--wallet, -w
The wallet parameter specifies the account wallet path for smart contract execution. Default: "./wallet.dat".

--account, -a
The account parameter specifies the account that will execute the contract.

--gasprice
The gasprice parameter specifies the gas price of the transfer transaction. The gas price of the transaction cannot be less than the lowest gas price set by node's transaction pool, otherwise the transaction will be rejected. The default value is 500 (0 in testmode). When there are transactions that are queued for packing into the block in the transaction pool, the transaction pool will deal with transactions according to the gas price and transactions with high gas prices will be prioritized.

--gaslimit
The gaslimit parameter specifies the gas limit of the transfer transaction. The gas limit of the transaction cannot be less than the minimum gas limit set by the node's transaction pool, otherwise the transaction will be rejected. Gasprice * gaslimit is actual ONG costs. The default value is 20000.

--address
The address parameter specifies the calling contract address.

--params
The params parameter is used to input the parameters of the contract invocation. The input parameters need to be encoded as described above.

--prepare, -p
The prepare parameter indicates that the current execution is a pre-executed contract. The transactions executed will not be packaged into blocks, nor will they consume any ONG. Pre-execution will return the contract method's return value, as well as the gas limit required for the current call.

--return
The return parameter is used with the --prepare parameter, which parses the return value of the contract by the return type of the --return parameter when the pre-execution is performed, otherwise returns the original value of the contract method call. Multiple return types are separated by "," such as string,int.


**Smart Contract Pre-Execution**

```
./Ontology contract invoke --address=XXX --params=XXX --return=XXX --p
```
Return example:

```
Contract invoke successfully
Gas consumed:20000
Return:0
```
**Smart Contract Execution**

```
./Ontology contract invoke --address=XXX --params=XXX --gaslimit=XXX
```

Before the smart contract is executed, the gas limit required by the current execution can be calculated through pre-execution to avoid execution failure due to insufficient ONG balance.

### 5.3 Smart Contract Code Execution Directly

Ontology supports direct execution of smart contact code after deploying a contract.

#### 5.3.1 Smart Contract Code Execution Directly Parameters

--wallet, -w
The wallet parameter specifies the account wallet path for smart contract execution. Default: "./wallet.dat".

--account, -a
The account parameter specifies the account that will execute the contract.

--gasprice
The gasprice parameter specifies the gas price of the transfer transaction. The gas price of the transaction cannot be less than the lowest gas price set by node's transaction pool, otherwise the transaction will be rejected. The default value is 500 (0 in testmode). When there are transactions that are queued for packing into the block in the transaction pool, the transaction pool will deal with transactions according to the gas price and transactions with high gas prices will be prioritized.

--gaslimit
The gaslimit parameter specifies the gas limit of the transfer transaction. The gas limit of the transaction cannot be less than the minimum gas limit set by the node's transaction pool, otherwise the transaction will be rejected. Gasprice * gaslimit is actual ONG costs. The default value is 20000.

--prepare, -p
The prepare parameter indicates that the current execution is a pre-executed contract. The transactions executed will not be packaged into blocks, nor will they consume any ONG. Pre-execution will return the contract method's return value, as well as the gas limit required for the current call.

--code
The code parameter specifies the code path of a smart contract.

**Smart Contract Code Execution Directly**

```
./Ontology contract invokeCode --code=XXX --gaslimit=XXX
```

## 6. Block Import and Export

Ontology CLI supports exporting the local node's block data to a compressed file. The generated compressed file can be imported into the Ontology node. For security reasons, the imported block data file must be obtained from a trusted source.

### 6.1 Export Blocks

#### 6.1.1 Export Block Parameters

--rpcport
The rpcport parameter specifies the port number to which the RPC server is bound. The default is 20336.

--exportfile
The exportfile parameter specifies the exported file path. The default value is: ./OntBlocks.dat.

--startheight
The startheight parameter specifies the start height of the exported block. Default value is 0.

--endheight
The endheight parameter specifies the end height of the exported block. When height of the local node's current block is greater than the end height required for export, the greater part will not be exported. Height is equal to 0, which means exporting all the blocks of the current node. The default value is 0.

--speed
The speed parameter specifies the export speed. Respectively, h denotes high, m denotes middle, and l denotes low. The default value is m.

Block export

```
./Ontology export
```

### 6.2 Import Blocks

#### 6.2.1 Importing Block Parameters

--data-dir
The data-dir parameter specifies the storage path of the block data. The default value is "./Chain".

--networkid
The networkid parameter is used to specify the network ID. Default value is 1, means MainNet network ID.

--config
The config parameter specifies the file path of the genesis block for the current Ontolgy node. Default value is main net config.

--disable-event-log
The disable-event-log parameter is used to disable the event log output when the smart contract is executed to improve the node transaction execution performance. The Ontology node enables the event log output function by default.

--endheight
The endheight parameter specifies the end height of the imported block. If the block height specified by --endheight is less than the maximum height of the block file, it will only be imported to the height specified by --endheight and the rest blocks will stop importing. The default value is 0, which means import all the blocks.

--importfile
The importfile parameter is used with --importfile to specify the path to the import file when importing blocks. The default value is "./OntBlocks.dat".

Import block

```
./ontology import --importfile=./OntBlocks.dat
```

## 7. Build Transaction

Build transaction command can build transaction raw data, such as transfer transaction, approve tansaction, and so on. Note that before send to Ontology, the transaction after built should be signed by private key.

### 7.1 Build Transfer Transaction

#### 7.1.1 Build Transfer Transaction Params

--wallet, -w
Wallet specifies the transfer-out account wallet path. The default value is: "./wallet.dat".

--gasprice
The gasprice parameter specifies the gas price of the transfer transaction. The gas price of the transaction cannot be less than the lowest gas price set by node's transaction pool, otherwise the transaction will be rejected. The default value is 500 (0 in testmode). When there are transactions that are queued for packing into the block in the transaction pool, the transaction pool will deal with transactions according to the gas price and transactions with high gas prices will be prioritized.

--gaslimit
The gaslimit parameter specifies the gas limit of the transfer transaction. The gas limit of the transaction cannot be less than the minimum gas limit set by the node's transaction pool, otherwise the transaction will be rejected. Gasprice * gaslimit is actual ONG costs. The default value is 20000.

--asset
The asset parameter specifies the asset type of the transfer. Ont indicates the ONT and ong indicates the ONG. The default value is ont.

--from
The from parameter specifies the transfer-out account address.

--to
The to parameter specifies the transfer-in account address.

--amount
The amount parameter specifies the transfer amount. Note: Since the precision of the ONT is 1, if the input is a floating-point value, then the value of the fractional part will be discarded; the precision of the ONG is 9, so the fractional part beyond 9 bits will be discarded.

--payer
payer parameter specifies the transaction fee payer. If don't specifies, using signer account default.

```
./ontology buildtx transfer --from=ARVVxBPGySL56CvSSWfjRVVyZYpNZ7zp48 --to=AaCe8nVkMRABnp5YgEjYZ9E5KYCxks2uce --amount=10
```

Return example:

```
Transfer raw tx:
00d1d376865bf401000000000000204e0000000000006a987e044e01e3b71f9bb60df57ab0458215ef0f6e00c66b6a146a987e044e01e3b71f9bb60df57ab0458215ef0fc86a14ca216237583e7c32ba82ca352ecc30782f5a902dc86a5ac86c51c1087472616e736665721400000000000000000000000000000000000000010068164f6e746f6c6f67792e4e61746976652e496e766f6b650000
```

### 7.2 Build Authorize Transfer

#### 7.2.1 Build Authorize Transfer Params

--wallet, -w
Wallet specifies the transfer-out account wallet path. The default value is: "./wallet.dat".

--gasprice
The gasprice parameter specifies the gas price of the transfer transaction. The gas price of the transaction cannot be less than the lowest gas price set by node's transaction pool, otherwise the transaction will be rejected. The default value is 500 (0 in testmode). When there are transactions that are queued for packing into the block in the transaction pool, the transaction pool will deal with transactions according to the gas price and transactions with high gas prices will be prioritized.

--gaslimit
The gaslimit parameter specifies the gas limit of the transfer transaction. The gas limit of the transaction cannot be less than the minimum gas limit set by the node's transaction pool, otherwise the transaction will be rejected. Gasprice * gaslimit is actual ONG costs. The default value is 20000.

--asset
The asset parameter specifies the asset type of the transfer. Ont indicates the ONT and ong indicates the ONG. The default value is ont.

--from
The from parameter specifies the transfer-out account address.

--to
The to parameter specifies the transfer-in account address.

--amount
The amount parameter specifies the transfer amount. Note: Since the precision of the ONT is 1, if the input is a floating-point value, then the value of the fractional part will be discarded; the precision of the ONG is 9, so the fractional part beyond 9 bits will be discarded.

--payer
payer parameter specifies the transaction fee payer. If don't specifies, using signer account default.

```
./ontology buildtx approve  --from=ARVVxBPGySL56CvSSWfjRVVyZYpNZ7zp48 --to=AaCe8nVkMRABnp5YgEjYZ9E5KYCxks2uce --amount=10
```
Return example:

```
Approve raw tx:
00d12178865bf401000000000000204e0000000000006a987e044e01e3b71f9bb60df57ab0458215ef0f6b00c66b6a146a987e044e01e3b71f9bb60df57ab0458215ef0fc86a14ca216237583e7c32ba82ca352ecc30782f5a902dc86a5ac86c07617070726f76651400000000000000000000000000000000000000010068164f6e746f6c6f67792e4e61746976652e496e766f6b650000
```
### 7.3 Build Transfer From Authorize Account Transaction

#### 7.3.1 Build Transfer From Authorize Account Transaction Params

--wallet, -w
Wallet specifies the wallet path of authorized account. The default value is: "./wallet.dat".

--gasprice
The gasprice parameter specifies the gas price of the transfer transaction. The gas price of the transaction cannot be less than the lowest gas price set by node's transaction pool, otherwise the transaction will be rejected. The default value is 500 (0 in testmode). When there are transactions that are queued for packing into the block in the transaction pool, the transaction pool will deal with transactions according to the gas price and transactions with high gas prices will be prioritized.

--gaslimit
The gaslimit parameter specifies the gas limit of the transfer transaction. The gas limit of the transaction cannot be less than the minimum gas limit set by the node's transaction pool, otherwise the transaction will be rejected. Gasprice * gaslimit is actual ONG costs. The default value is 20000.

--asset
The asset parameter specifies the asset type of the transfer. Ont indicates the ONT and ong indicates the ONG. The default value is ont.

--from
The from parameter specifies the transfer-out account address.

--to
The to parameter specifies the transfer-in account address.

--sender
The sender parameter specifies the account address that actually operates the authorized transfer. If no sender parameter is specified, the sender parameter defaults to the value of the to parameter.

--amount
The amount parameter specifies the transfer amount and the transfer amount cannot be greater than the authorized transfer balance. Otherwise, the transaction will be rejected. Note: Since the precision of the ONT is 1, if the input is a floating-point value, then the value of the fractional part will be discarded; the precision of the ONG is 9, so the fractional part beyond 9 bits will be discarded.

--payer
payer parameter specifies the transaction fee payer. If don't specifies, using signer account default.

```
./ontology buildtx transferfrom --sender=AMFrW7hrSRw1Azz6hQohni8BdStZDvectW --from=Aaxjf7utmjSstmTD1LjtYfhZ3CoWaxC7Tt --to=AMFrW7hrSRw1Azz6hQohni8BdStZDvectW --amount=10
```

Return example:

```
00d10754875bf401000000000000204e0000000000003c2352095b7428debfd1c1519f5a8f45a474a4218700c66b6a143c2352095b7428debfd1c1519f5a8f45a474a421c86a14d2784bddeac73d20124f20f4fa9528f3365a4dd4c86a143c2352095b7428debfd1c1519f5a8f45a474a421c86a5ac86c0c7472616e7366657246726f6d1400000000000000000000000000000000000000010068164f6e746f6c6f67792e4e61746976652e496e766f6b650000
```

### 7.4 Build Withdraw ONG Transaction

#### 7.4.1 Build Withdraw ONG Transaction Params

--wallet, -w
Wallet specifies the wallet path of authorized account. The default value is: "./wallet.dat".

--gasprice
The gasprice parameter specifies the gas price of the transfer transaction. The gas price of the transaction cannot be less than the lowest gas price set by node's transaction pool, otherwise the transaction will be rejected. The default value is 500 (0 in testmode). When there are transactions that are queued for packing into the block in the transaction pool, the transaction pool will deal with transactions according to the gas price and transactions with high gas prices will be prioritized.

--gaslimit
The gaslimit parameter specifies the gas limit of the transfer transaction. The gas limit of the transaction cannot be less than the minimum gas limit set by the node's transaction pool, otherwise the transaction will be rejected. Gasprice * gaslimit is actual ONG costs. The default value is 20000.

--amount
The amount parameter specifies the transfer amount and the transfer amount cannot be greater than the authorized transfer balance. Otherwise, the transaction will be rejected. Note: Since the precision of the ONT is 1, if the input is a floating-point value, then the value of the fractional part will be discarded; the precision of the ONG is 9, so the fractional part beyond 9 bits will be discarded.
If don't specifies withdraw all unbound ONG as default.

--receive
receive params specifies the ONG receive account. If don't specifies, using withdraw account default.

--payer
payer parameter specifies the transaction fee payer. If don't specifies, using signer account default.

--rpcport
The rpcport parameter specifies the port number to which the RPC server is bound. The default is 20336.

```
./ontology buildtx withdrawong ARVVxBPGySL56CvSSWfjRVVyZYpNZ7zp48
```

Return example:

```
Withdraw account:ARVVxBPGySL56CvSSWfjRVVyZYpNZ7zp48
Receive account:ARVVxBPGySL56CvSSWfjRVVyZYpNZ7zp48
Withdraw ONG amount:2321499191858975
Withdraw raw tx:
00d11b56875bf401000000000000204e0000000000006a987e044e01e3b71f9bb60df57ab0458215ef0f8e00c66b6a146a987e044e01e3b71f9bb60df57ab0458215ef0fc86a140000000000000000000000000000000000000001c86a146a987e044e01e3b71f9bb60df57ab0458215ef0fc86a071f57ad26643f08c86c0c7472616e7366657246726f6d1400000000000000000000000000000000000000020068164f6e746f6c6f67792e4e61746976652e496e766f6b650000
```

## 8. Sign To Transaction

The transaction build by buildtx command, should be signed before send to Ontology. Note that if transction fee payer is different with transfer from, both account should sign to the transaction.

### 8.1 Sign To Transaction Parameters

--wallet, -w
Wallet specifies the wallet path of authorized account. The default value is: "./wallet.dat".

--account, a
account parameter specifies signature account, if not specified, the default account of wallet will be used.

--send
--send parameter specifies whether send transaction to Ontology after signed.

--prepare
prepare parameter specifies whether prepare execute transaction, without send to Ontology.

--rpcport
The rpcport parameter specifies the port number to which the RPC server is bound. The default is 20336.

```
./ontology sigtx --account=ARVVxBPGySL56CvSSWfjRVVyZYpNZ7zp48 00d11b56875bf401000000000000204e0000000000006a987e044e01e3b71f9bb60df57ab0458215ef0f8e00c66b6a146a987e044e01e3b71f9bb60df57ab0458215ef0fc86a140000000000000000000000000000000000000001c86a146a987e044e01e3b71f9bb60df57ab0458215ef0fc86a071f57ad26643f08c86c0c7472616e7366657246726f6d1400000000000000000000000000000000000000020068164f6e746f6c6f67792e4e61746976652e496e766f6b650000
```

Return example:

```
RawTx after signed:
00d11b56875bf401000000000000204e0000000000006a987e044e01e3b71f9bb60df57ab0458215ef0f8e00c66b6a146a987e044e01e3b71f9bb60df57ab0458215ef0fc86a140000000000000000000000000000000000000001c86a146a987e044e01e3b71f9bb60df57ab0458215ef0fc86a071f57ad26643f08c86c0c7472616e7366657246726f6d1400000000000000000000000000000000000000020068164f6e746f6c6f67792e4e61746976652e496e766f6b65000141407331b7ba2a7708187ad4cb14146d2080185e42f0a39d572f58d25fa2e20f3066711b64f2b91d958683f7bfb904badeb0d6bc733506e665028a2c2968b77d5958232103c0c30f11c7fc1396e8595bf2e339d553d728ea6f21ae831e8ab704ca14fe8a56ac
```

## 9. Generate Multi-Signature Address

Generating a multi-signature address need public keys and the signature number at least.

### 9.1 Generate Multi-Signature Address Parameters

--pubkey
pubkey parameter specifies the range of signature of public key.
The public key of account can get by the command:

```
./ontology account list -v
```
The max number of public key support by Ontology is 16 at present.

-m
m parameter specifies the least number of signature. Default value is 1.

```
./ontology multisigaddr --pubkey=03c0c30f11c7fc1396e8595bf2e339d553d728ea6f21ae831e8ab704ca14fe8a56,02b2b9fb60a0add9ef6715ffbac8bc7e81cb47cd06c157c19e6a858859c0158231 -m=1
```
Return example:

```
Pub key list:
Index 1 Address:AaCe8nVkMRABnp5YgEjYZ9E5KYCxks2uce PubKey:02b2b9fb60a0add9ef6715ffbac8bc7e81cb47cd06c157c19e6a858859c0158231
Index 2 Address:ARVVxBPGySL56CvSSWfjRVVyZYpNZ7zp48 PubKey:03c0c30f11c7fc1396e8595bf2e339d553d728ea6f21ae831e8ab704ca14fe8a56

MultiSigAddress:Ae4cxJiubmgueAVtNbjpmm2AGNgdKP6Ea7
```

## 10. Multi-Signature To Transaction

Multi-Signature need multiple account sign to transaction one by one. The output of signature as input of another signature, until signature number up to M.

### 10.1 Multi-signature parameters for transactions

--pubkey
is used to specify a list of public keys for multiple signatures, separated by a comma ','. The account public key can be passed through the command:

pubkey parameter specifies the range of signature of public key.
The public key of an account can be got by the command:

```
./ontology account list -v
```
The max number of public key support by Ontology is 16 at present.

-m
m parameter specifies the least number of signature. Default value is 1.

--wallet, -w
Wallet specifies the wallet path of authorized account. The default value is: "./wallet.dat".

--account, a
account parameter specifies signature account, if not specified, the default account of wallet will be used.

--send
--send parameter specifies whether send transaction to Ontology after signed.

--prepare
prepare parameter specifies whether prepare execute transaction, without send to Ontology.

--rpcport
The rpcport parameter specifies the port number to which the RPC server is bound. The default is 20336.

```
./ontology multisigtx --account=ARVVxBPGySL56CvSSWfjRVVyZYpNZ7zp48 --pubkey=03c0c30f11c7fc1396e8595bf2e339d553d728ea6f21ae831e8ab704ca14fe8a56,02b2b9fb60a0add9ef6715ffbac8bc7e81cb47cd06c157c19e6a858859c0158231 -m=1 00d1045f875bf401000000000000204e000000000000f47d92d27d02b93d21f8af16c9f05a99d128dd5a6e00c66b6a14f47d92d27d02b93d21f8af16c9f05a99d128dd5ac86a14ca216237583e7c32ba82ca352ecc30782f5a902dc86a5ac86c51c1087472616e736665721400000000000000000000000000000000000000010068164f6e746f6c6f67792e4e61746976652e496e766f6b650000
```

Return example:

```
RawTx after multi signed:
00d1045f875bf401000000000000204e000000000000f47d92d27d02b93d21f8af16c9f05a99d128dd5a6e00c66b6a14f47d92d27d02b93d21f8af16c9f05a99d128dd5ac86a14ca216237583e7c32ba82ca352ecc30782f5a902dc86a5ac86c51c1087472616e736665721400000000000000000000000000000000000000010068164f6e746f6c6f67792e4e61746976652e496e766f6b65000141409dd2a46277f96566b9e9b4fc354be90b61776c58125cfbf36e770b1b1d50a16febad4bfadfc966fa575e90acf3b8308d7a0f637260b31321cb7ef6f741364d0e47512102b2b9fb60a0add9ef6715ffbac8bc7e81cb47cd06c157c19e6a858859c01582312103c0c30f11c7fc1396e8595bf2e339d553d728ea6f21ae831e8ab704ca14fe8a5652ae
```

## 11. Send Transaction

The transaction after being signed can be sent to Ontology via sendtx command.

### 11.1 Send Transaction Parameters

--rpcport
The rpcport parameter specifies the port number to which the RPC server is bound. The default is 20336.

--prepare
prepare parameter specifies whether prepare execute transaction, without send to Ontology.

```
./ontology sendtx 00d17c61875bf401000000000000204e0000000000006a987e044e01e3b71f9bb60df57ab0458215ef0f6e00c66b6a146a987e044e01e3b71f9bb60df57ab0458215ef0fc86a14ca216237583e7c32ba82ca352ecc30782f5a902dc86a5ac86c51c1087472616e736665721400000000000000000000000000000000000000010068164f6e746f6c6f67792e4e61746976652e496e766f6b65000141409f32f1fd170d174959da26cb9df8f4a15049d255ed3953d92870d5739c4e8b8158ec3bde1e9ae9b4d9621b09311b5e49ed91dcbc64d3b5f74cf011eaa616c403232103c0c30f11c7fc1396e8595bf2e339d553d728ea6f21ae831e8ab704ca14fe8a56ac
```

Return example:

```
  TxHash:f8ea91da985af249e808913b6398150079cdfb02273146e4eb69c43947a42db2

Tip:
  Using './ontology info status f8ea91da985af249e808913b6398150079cdfb02273146e4eb69c43947a42db2' to query transaction status.
```

If prepare execute return:

```
Prepare execute transaction success.
Gas limit:20000
Result:01
```

## 12. Show Transaction Infomation

The information of transaction field can be show via showtx command.

Example:

```
./ontology showtx 00d1045f875bf401000000000000204e000000000000f47d92d27d02b93d21f8af16c9f05a99d128dd5a6e00c66b6a14f47d92d27d02b93d21f8af16c9f05a99d128dd5ac86a14ca216237583e7c32ba82ca352ecc30782f5a902dc86a5ac86c51c1087472616e736665721400000000000000000000000000000000000000010068164f6e746f6c6f67792e4e61746976652e496e766f6b65000141409dd2a46277f96566b9e9b4fc354be90b61776c58125cfbf36e770b1b1d50a16febad4bfadfc966fa575e90acf3b8308d7a0f637260b31321cb7ef6f741364d0e47512102b2b9fb60a0add9ef6715ffbac8bc7e81cb47cd06c157c19e6a858859c01582312103c0c30f11c7fc1396e8595bf2e339d553d728ea6f21ae831e8ab704ca14fe8a5652ae
```
Return:

```
{
   "Version": 0,
   "Nonce": 1535598340,
   "GasPrice": 500,
   "GasLimit": 20000,
   "Payer": "Ae4cxJiubmgueAVtNbjpmm2AGNgdKP6Ea7",
   "TxType": 209,
   "Payload": {
      "Code": "00c66b6a14f47d92d27d02b93d21f8af16c9f05a99d128dd5ac86a14ca216237583e7c32ba82ca352ecc30782f5a902dc86a5ac86c51c1087472616e736665721400000000000000000000000000000000000000010068164f6e746f6c6f67792e4e61746976652e496e766f6b65",
      "GasLimit": 0
   },
   "Attributes": [],
   "Sigs": [
      {
         "PubKeys": [
            "02b2b9fb60a0add9ef6715ffbac8bc7e81cb47cd06c157c19e6a858859c0158231",
            "03c0c30f11c7fc1396e8595bf2e339d553d728ea6f21ae831e8ab704ca14fe8a56"
         ],
         "M": 1,
         "SigData": [
            "9dd2a46277f96566b9e9b4fc354be90b61776c58125cfbf36e770b1b1d50a16febad4bfadfc966fa575e90acf3b8308d7a0f637260b31321cb7ef6f741364d0e"
         ]
      }
   ],
   "Hash": "34559b63187d7ddf5a17ac7a2dabb8fcaa1bea6676eba78a174d038ff3c66f15",
   "Height": 0
}
```
