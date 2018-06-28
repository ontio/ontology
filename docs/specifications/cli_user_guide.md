# Ontology CLI Instruction

Ontology CLI is an Ontology node command line Client for starting and managing Ontology nodes, managing user wallets, sending transactions, and deploying and invoking contracts.

## 1. Start and Manage Ontology Nodes

Ontology CLI has a lot of startup parameters for configuring some of the Ontology node's behavior. Use ./Ontology -help to see all startup parameters supported by the Ontology CLI node. If Ontology CLI is started without any parameters, it will access the Ontology Polaris test network as a synchronous node by default.

### 1.1 Startup Parameters

The following are the command line parameters supported by Ontology CLI:

#### 1.1.1 Ontology System Parameters

--config
The config parameter specifies the file path of the genesis block for the current Ontolgy node. If doesn't specify, Ontology will use the config of Polaris test net. Note that the genesis block configuration must be the same for all nodes in the same network, otherwise it will not be able to synchronize blocks or start nodes due to block data incompatibility.

--loglevel
The loglevel parameter is used to set the log level the Ontology outputs. Ontology supports 7 different log levels, i.e. 0:Debug 1:Info 2:Warn 3:Error 4:Fatal 5:Trace 6:MaxLevel. The logs are logged from low to high, and the log output volume is from high to low. The default value is 1, which means that only logs at the info level or higher level.

--disableeventlog
The disableeventlog parameter is used to disable the event log output when the smart contract is executed to improve the node transaction execution performance. The Ontology node enables the event log output function by default.

--datadir
The datadir parameter specifies the storage path of the block data. The default value is "./Chain".

--import
The import parameter is used to start the block import function of the Ontology node and increase the block synchronization speed by importing local files.


--importheight
The importheight parameter is used with --import to specify the height of the imported target block. If the height of the block specified by importheight is less than the maximum height of the block file, it will import all blocks whose height is less than the importheight value. The default value is 0, which means import all the blocks.


--importfile
The importfile parameter is used with --import to specify the imported file path. The default value is "./blocks.dat".

#### 1.1.2 Account Parameters

--wallet, -w
The wallet parameter is used to specify the wallet file path when the Ontology node starts. The default value is "./wallet.dat".

--account, -a
The account parameter is used to specify the account address when the Ontology node starts. If the account is null, it uses the wallet default account.

--password, -p
The password parameter is used to specify the account password when Ontology node starts. Because the account password entered in the command line is saved in the log, it is easy to leak the password. Therefore, it is not recommended to use this parameter in a production environment.

#### 1.1.3 Consensus Parameters

--enableconsensus
The enableconsensus parameter is used to turn the consensus on. If the current node will startup as a bookkeeper node, must enable this flag. The default is disable.

--maxtxinblock
The maxtxinblock parameter is used to set the maximum transaction number of a block. The default value is 50000.

#### 1.1.4 P2P Network Parameters

--networkid
The networkid parameter is used to specify the network ID. Different networkids cannot connect to the blockchain network. 1=main net, 2=polaris test net, 3=testmode, and other for custom network.

--nodeport
The nodeport parameter is used to specify the P2P network port number. The default value is 20338.

--consensusport
The consensusport parameter specifies the consensus network port number. By default, the consensus network reuses the P2P network, so it is not necessary to specify a consensus network port. After the dual network is enabled with the --dualport parameter, the consensus network port number must be set separately. The default is 20339.

--dualport
The dualport parameter initiates a dual network, i.e. a P2P network for processing transaction messages and a consensus network for consensus messages. The parameter disables by default.


#### 1.1.5 RPC Server Parameters

--disablerpc
The disablerpc parameter is used to shut down the rpc server. The Ontology node starts the rpc server by default at startup.

--rpcport
The rpcport parameter specifies the port number to which the rpc server is bound. The default is 20336.

#### 1.1.6 Restful Server Parameters

--rest
The rest parameter is used to start the rest server.

--restport
The restport parameter specifies the port number to which the restful server is bound. The default value is 20334.

#### 1.1.7 Web Socket Server Parameters

--ws
The ws parameter is used to start the Web socket server.

--wsport
The wsport parameter specifies the port number to which the web socket server is bound. The default value is 20335

#### 1.1.8 Test Mode Parameters

--testmode
The testmode parameter is used to start a single node test network for ease of development and debug. In testmode, Ontology will start rpc, rest and web socket server, and block chain data will be clear generated by last start in testmode.

--testmodegenblocktime
The testmodegenblocktime parameter is used to set the block-out time in test mode. The time unit is in seconds, and the minimum block-out time is 2 seconds.

#### 1.1.9 Transaction Parameter

--gasprice
The gasprice parameter is used to set the lowest gasprice of the current node transaction pool to accept transactions. Transactions below this gasprice will be discarded. The default value is 0.

--gaslimit
The gaslimit parameter is used to set the gaslimit of the current node transaction pool to accept transactions. Transactions below this gaslimit will be discarded. The default value is 30000.

### 1.2 Node Deployment

#### 1.2.1 Genesis Block Configuration File

For deploying node, we need to provide the genesis block configuration file for initializing the genesis block. Ontology supports the VBFT and dBFT consensus algorithms. A blockchain network must use the same consensus algorithm, otherwise nodes will not work properly. Because of Ontology  built-in the config of Polaris test net, if you start Ontology without specifies any  config of genesis, will access Polaris test net. However, if you need to build an Ontology private chain, you need to provide the config of genesis block via --config flag.

To deploy Ontology test network with dBFT consensus algorithm requires a minimum of 4 nodes, and to deploy Ontology test network with the VBFT consensus algorithm requires a minimum of 7 nodes.

##### 1.2.1.1 VBFT Configuration File

An example of a genesis block configuration file using the VBFT consensus algorithm:

```json
{
  "SeedList": [
    "192.168.0.1:20338"             //Seed node list
  ],
  "ConsensusType":"vbft",           //Specify the current network to use VBFT algorithm
  "VBFT":{                          //VBFT algorithm configuration
    "n":7,                          //Network size (temporarily unused)
    "c":2,                          //The number of fault-tolerant nodes
    "k":7,                          //The number of consensus nodes
    "l":112,                        //Length of POS table
    "block_msg_delay":10000,        //Maximum broadcast delay for block messages (ms)
    "hash_msg_delay":10000,         //Maximum broadcast delay for hash messages (ms)
    "peer_handshake_timeout":10,    //Node handshake timeout (s)
    "max_block_change_view":1000,   //Consensus period
    "vrf_value":"1c9810aa9822e511d5804a9c4db9dd08497c31087b0daafa34d768a3253441fa20515e2f30f81741102af0ca3cefc4818fef16adb825fbaa8cad78647f3afb590e",
    "vrf_proof":"c57741f934042cb8d8b087b44b161db56fc3ffd4ffb675d36cd09f83935be853d8729f3f5298d12d6fd28d45dde515a4b9d7f67682d182ba5118abf451ff1988",
    "peers":[                       //Bookkeeping node configuration
      {
        "index":1,                  //Bookkeeping node index
        "peerPubkey":"1202028541d32f3b09180b00affe67a40516846c16663ccb916fd2db8106619f087527",  //Bookkeeping node public key
        "address":"TA9TVuR4Ynn4VotfpExY5SaEy8a99obFPr",                                         //Bookkeeping node account address
        "initPos":1000              // initial ONT mortgage value
      },
      {
        "index":2,
        "peerPubkey":"120202dfb161f757921898ec2e30e3618d5c6646d993153b89312bac36d7688912c0ce",
        "address":"TA9TVuR4Ynn4VotfpExY5SaEy8a99obFPr",
        "initPos":2000
      },
      {
        "index":3,
        "peerPubkey":"1202039dab38326268fe82fb7967fe2e7f5f6eaced6ec711148a66fbb8480c321c19dd",
        "address":"TA9TVuR4Ynn4VotfpExY5SaEy8a99obFPr",
        "initPos":3000
      },
      {
        "index":4,
        "peerPubkey":"12020384f2729bc5d9b14dcbf17aba108261dc7ad867127e413d3c8bfb4731739687b3",
        "address":"TA9TVuR4Ynn4VotfpExY5SaEy8a99obFPr",
        "initPos":4000
      },
      {
        "index":5,
        "peerPubkey":"120203362f99284daa9f581fab596516f75475fc61a5f80de0e268a68430dc7589859c",
        "address":"TA9TVuR4Ynn4VotfpExY5SaEy8a99obFPr",
        "initPos":3000
      },
      {
        "index":6,
        "peerPubkey":"120203db6e37a2d897f2d61b42dcd478323a8a20c3444af4ee29653849f38d0bdb67f4",
        "address":"TA9TVuR4Ynn4VotfpExY5SaEy8a99obFPr",
        "initPos":2000
      },
      {
        "index":7,
        "peerPubkey":"12020298fe9f22e9df64f6bfcc1c2a14418846cffdbbf510d261bbc3fa6d47073df9a2",
        "address":"TA9TVuR4Ynn4VotfpExY5SaEy8a99obFPr",
        "initPos":1000
      }
    ]
  }
}
```

##### 1.2.1.2 dBFT Configuration File

An example of a genesis block configuration file using the dBFT consensus algorithm:

```json
{
  "SeedList": [
    "192.168.0.1:20338"               //Seed node list
  ],
  "ConsensusType":"dbft",             //Specify the current network to use dBFT algorithm
  "DBFT":{                            //dBFT consensus algorithm configuration
    "Bookkeepers": [                  //Bookkeeping node public key list
      "1202028541d32f3b09180b00affe67a40516846c16663ccb916fd2db8106619f087527",
      "120202dfb161f757921898ec2e30e3618d5c6646d993153b89312bac36d7688912c0ce",
      "1202039dab38326268fe82fb7967fe2e7f5f6eaced6ec711148a66fbb8480c321c19dd",
      "12020384f2729bc5d9b14dcbf17aba108261dc7ad867127e413d3c8bfb4731739687b3"
    ],
    "GenBlockTime":6                  //block-out time interval, in seconds
  }
}
```

#### 1.2.2 Bookkeeping Node Deployment

According to different roles of nodes, they can be divided into bookkeeping nodes and synchronization nodes. Bookkeeping nodes participate in the network consensus, and synchronization nodes only synchronize the blocks generated by the bookkeeping nodes. Since Ontology node wont't start consensus by default, must turn consensus on by the --enableconsensus parameter. The Ontology node will start the Rpc server by default and output the Event Log of the smart contract. Therefore, if there is no special requirement, you can use the --disablerpc and --disableeventlog command line parameters to turn off the rpc and eventlog modules.

Recommended bookkeeping node startup parameters:

```
./Ontology --enbaleconsensus --disablerpc --disableeventlog
```

If the node does not use the default genesis block configuration file and wallet account, the node can specify them with the --config, --wallet, --account parameters.
At the same time, if the bookkeeping node needs to modify the default minimum gas price and gas limit of the transaction pool, it can set the parameters by --gasprice and --gaslimit.

#### 1.2.3 Synchronization Node Deployment

Since the synchronization node only synchronizes the blocks generated by the bookkeeping node and does not participate in the network consensus.

```
./Ontology
```

If the node does not use the default genesis block configuration file, it can be specified with the --config parameter. Since wont't turn consensus on, don't need wallet when startup a synchronization node.

#### 1.2.4 Deploying a Multi-node Test Network on a Single Computer

You can deploy Ontology's test network on a single machine, which is essentially no different from deploying multiple nodes on multiple machines, but note that you need to use different port numbers:

```
./Ontology --nodeport=XXX --rpcport=XXX
```
#### 1.2.5 Single-Node Test Network Deployment

Ontology supports single-node network deployment for the development of test environments. To start a single-node test network, you only need to add the --testmode command line parameter.

```
./Ontology --testmode
```

If the node does not use the default genesis block configuration file and wallet account, the node can specify them with the --config, --wallet, --account parameters.
At the same time, if the bookkeeping node needs to modify the default minimum gas price and gas limit of the transaction pool, it can set the parameters by --gasprice and --gaslimit.

Note that, Ontology will turn consensus rpc, rest and web socket server on in testmode.

## 2. Wallet Management

Wallet management commands can be used to add, view, modify, delete, import account.
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

**Add account**

```
./Ontology account add --default
```

You can view the help information by ./Ontology account add --help.

### 2.2. View Account

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

--changepasswd
The changepasswd parameter is used to modify the account password.

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
The source parameter specifies the imported wallet path

```
./Ontology account import -s=./source_wallet.dat
```

## 3. Asset Management

Asset management commands can check account balance, ONT/ONG transfers, extract ONGs, and view unbound ONGs.

### 3.1 Check Your Account Balance

```
./Ontology asset balance <address|index|label>
```
### 3.2 ONT/ONG Transfers

#### 3.2.1 Transfer Arguments

--wallet, -w
Wallet specifies the transfer-out account wallet path. The default value is: "./wallet.dat".

--gasprice
The gasprice parameter specifies the gas price of the transfer transaction. The gas price of the transaction cannot be less than the lowest gas price set by node's transaction pool, otherwise the transaction will be rejected. The default value is 0. When there are transactions that are queued for packing into the block in the transaction pool, the transaction pool will deal with transactions according to the gas price and transactions with high gas prices will be prioritized.

--gaslimit
The gaslimit parameter specifies the gas limit of the transfer transaction. The gas limit of the transaction cannot be less than the minimum gas limit set by the node's transaction pool, otherwise the transaction will be rejected. Gasprice * gaslimit is actual ONG costs. The default value is 30000.

--asset
The asset parameter specifies the asset type of the transfer. Ont indicates the ONT and ong indicates the ONG. The default value is ont.

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
The gasprice parameter specifies the gas price of the transfer transaction. The gas price of the transaction cannot be less than the lowest gas price set by node's transaction pool, otherwise the transaction will be rejected. The default value is 0. When there are transactions that are queued for packing into the block in the transaction pool, the transaction pool will deal with transactions according to the gas price and transactions with high gas prices will be prioritized.

--gaslimit
The gaslimit parameter specifies the gas limit of the transfer transaction. The gas limit of the transaction cannot be less than the minimum gas limit set by the node's transaction pool, otherwise the transaction will be rejected. Gasprice * gaslimit is actual ONG costs. The default value is 30000.

--asset
The asset parameter specifies the asset type of the transfer. Ont indicates the ONT and ong indicates the ONG. The default value is ont.

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
The asset parameter specifies the asset type of the transfer. Ont indicates the ONT and ong indicates the ONG. The default value is ont.

--from
The from parameter specifies the transfer-out account address.

--to
The to parameter specifies the transfer-in account address.

**View Authorized Transfer Balance**

```
./Ontology asset allowance --from=<address|index|label> --to=<address|index|label>
```

### 3.5 Transferring Fund from Authorized Accounts

After user authorization, the transfer can be made from the authorized account.

#### 3.5.1 Transferring Fund from Authorized Accounts Parameters
--wallet, -w
Wallet specifies the wallet path of authorized account. The default value is: "./wallet.dat".

--gasprice
The gasprice parameter specifies the gas price of the transfer transaction. The gas price of the transaction cannot be less than the lowest gas price set by node's transaction pool, otherwise the transaction will be rejected. The default value is 0. When there are transactions that are queued for packing into the block in the transaction pool, the transaction pool will deal with transactions according to the gas price and transactions with high gas prices will be prioritized.

--gaslimit
The gaslimit parameter specifies the gas limit of the transfer transaction. The gas limit of the transaction cannot be less than the minimum gas limit set by the node's transaction pool, otherwise the transaction will be rejected. Gasprice * gaslimit is actual ONG costs. The default value is 30000.

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

**Transfer from authorized account**

```
./Ontology asset transferfrom --from=<address|index|label> --to=<address|index|label> --sender=<address|index|label> --amount=XXX
```

### 3.6 View Unlocked ONG Balance

The ONG adopts the timing unbundling policy to release the ONG bound to the ONT. Use the following command to view the current account as the unlocked ONG balance.

```
./Ontology asset unboundong <address|index|label>
```

### 3.7 Extract Unlocked ONG

Use the following command to extract all unlocked ONG.

#### 3.7.1 Extracting Unlocked ONG Parameters

--wallet, -w
Wallet specifies the wallet path of extracted account. The default value is: "./wallet.dat".

--gasprice
The gasprice parameter specifies the gas price of the transfer transaction. The gas price of the transaction cannot be less than the lowest gas price set by node's transaction pool, otherwise the transaction will be rejected. The default value is 0. When there are transactions that are queued for packing into the block in the transaction pool, the transaction pool will deal with transactions according to the gas price and transactions with high gas prices will be prioritized.

--gaslimit
The gaslimit parameter specifies the gas limit of the transfer transaction. The gas limit of the transaction cannot be less than the minimum gas limit set by the node's transaction pool, otherwise the transaction will be rejected. Gasprice * gaslimit is actual ONG costs. The default value is 30000.

**Extract Unlocked ONG**
```
./Ontology asset withdrawong <address|index|label>
```
## 4 Query Information

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
Among them, State represents the execution result of the transaction. The value of State is 1, indicating that the transaction execution is successful. When the State value is 0, it indicates that the execution failed. GasConsumed indicates the ONG consumed by the transaction execution. Notify represents the Event log output when the transaction is executed. Different transactions may output different Event logs.

## 5. Smart Contract

Smart contract operations support the deployment of NeoVM smart contract, and the pre-execution and execution of NeoVM smart contract.

### 5.1 Smart Contract Deployment

Before smart deployment, you need to compile the NeoVM contract compiler such as [SmartX] (http://smartx.ont.io) and save the compiled code in a local text file.

#### 5.1.1 Smart Contract Deployment Parameters

--wallet, -w
The wallet parameter specifies the wallet path of account for deploying smart contracts. Default: "./wallet.dat".

--account, -a
The account parameter specifies the account that a contract deploys.

--gasprice
The gasprice parameter specifies the gas price of the transfer transaction. The gas price of the transaction cannot be less than the lowest gas price set by node's transaction pool, otherwise the transaction will be rejected. The default value is 0. When there are transactions that are queued for packing into the block in the transaction pool, the transaction pool will deal with transactions according to the gas price and transactions with high gas prices will be prioritized.

--gaslimit
The gaslimit parameter specifies the gas limit of the transfer transaction. The gas limit of the transaction cannot be less than the minimum gas limit set by the node's transaction pool, otherwise the transaction will be rejected. Gasprice * gaslimit is actual ONG costs. The default value is 30000.

**For contract deployments, the gaslimit value must be greater than 10000000, and there must be sufficient ONG balance in the account.**

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
The gasprice parameter specifies the gas price of the transfer transaction. The gas price of the transaction cannot be less than the lowest gas price set by node's transaction pool, otherwise the transaction will be rejected. The default value is 0. When there are transactions that are queued for packing into the block in the transaction pool, the transaction pool will deal with transactions according to the gas price and transactions with high gas prices will be prioritized.

--gaslimit
The gaslimit parameter specifies the gas limit of the transfer transaction. The gas limit of the transaction cannot be less than the minimum gas limit set by the node's transaction pool, otherwise the transaction will be rejected. Gasprice * gaslimit is actual ONG costs. The default value is 30000.

--address
The address parameter specifies the calling contract address

--params
The params parameter is used to input the parameters of the contract invocation. The input parameters need to be encoded as described above.

--prepare, -p
The prepare parameter indicates that the current execution is a pre-executed contract. The transactions executed will not be packaged into blocks, nor will they consume any ONG. Pre-execution will return the contract method's return value, as well as the gas limit required for the current call.

--return
The return parameter is used with the --prepare parameter, which parses the return value of the contract by the return type of the --return parameter when the pre-execution is performed, otherwise returns the original value of the contract method call. Multiple return types are separated by "," such as string,int


**Smart Contract Pre-Execution**

```
./Ontology contract invoke --address=XXX --params=XXX --return=XXX --p
```
Return example：

```
Contract invoke successfully
Gas consumed:30000
Return:0
```
**Smart Contract Execution**

```
./Ontology contract invoke --address=XXX --params=XXX --gaslimit=XXX
```

Before the smart contract is executed, the gas limit required by the current execution can be calculated through pre-execution to avoid execution failure due to insufficient ONG balance.

### 5.3 Smart Contract Code Execution Directly

Ontology support execut smart contact code directly, after deploy contract.

#### 5.3.1 Smart Contract Code Execution Directly Parameters

--wallet, -w
The wallet parameter specifies the account wallet path for smart contract execution. Default: "./wallet.dat".

--account, -a
The account parameter specifies the account that will execute the contract.

--gasprice
The gasprice parameter specifies the gas price of the transfer transaction. The gas price of the transaction cannot be less than the lowest gas price set by node's transaction pool, otherwise the transaction will be rejected. The default value is 0. When there are transactions that are queued for packing into the block in the transaction pool, the transaction pool will deal with transactions according to the gas price and transactions with high gas prices will be prioritized.

--gaslimit
The gaslimit parameter specifies the gas limit of the transfer transaction. The gas limit of the transaction cannot be less than the minimum gas limit set by the node's transaction pool, otherwise the transaction will be rejected. Gasprice * gaslimit is actual ONG costs. The default value is 30000.

--prepare, -p
The prepare parameter indicates that the current execution is a pre-executed contract. The transactions executed will not be packaged into blocks, nor will they consume any ONG. Pre-execution will return the contract method's return value, as well as the gas limit required for the current call.

--code
The code parameter specifies the code path of a smart contract.

**Smart Contract Code Execution Directly**

```
./Ontology contract invokeCode --code=XXX --gaslimit=XXX
```

## 6、Block Import and Export

Ontology CLI supports exporting the local node's block data to a compressed file. The generated compressed file can be imported into the Ontology node. For security reasons, the imported block data file must be obtained from a trusted source.

### 6.1 Export Blocks

#### 6.1.1 Export Block Parameters

--file
The file parameter specifies the exported file path. The default value is: blocks.dat

--height
The height parameter specifies the height of the exported block. When height of the local node's current block is greater than the height required for export, the greater part will not be exported. Height is equal to 0, which means exporting all the blocks of the current node. The default value is 0.

--speed
The speed parameter specifies the export speed. Respectively, h denotes high, m denotes middle, and l denotes low. The default value is m.

Block export

```
./Ontology export
```

### 6.2 Import Blocks

#### 6.2.1 Importing Block Parameters

--importheight
The importheight parameter specifies the height of the imported target block. If the block height specified by importheight is less than the maximum height of the block file, it will only be imported to the height specified by importheight and the rest blocks will stop importing. The default value is 0, which means import all the blocks.

--importfile
The importfile parameter is used with --import to specify the path to the import file when importing blocks. The default value is "./blocks.dat".

Import block

```
./Ontology import
```
