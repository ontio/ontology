# Onchain DNA Data Model Specification

* [Introduction](#introduction)
* [Definitions](#definitions)
  * [Ledger](#ledger)
  * [Blockchain](#Blockchain)
  * [Block](#Block)
  * [Blockheader](#Blockheader)
  * [Transaction](#Transaction)
     * [Payload](#Payload)
  * [Contract](#Contract)
  * [SignableData](#SignableData)
  * [ContractContext](#ContractContext)
  

## Introduction

This document describes the serialization format for the data structures used in the Onchain DNA.

## Definitions

### Ledger

The ledger Contains the BlockChain, current state and the store interface of the ledger. which is maintained by each peer.
DNA is designed to support muti Ledger, but now is only allowed one ledger.


Field               | Type              | Description
--------------------|-------------------|----------------------------------------------------------
Blockchain          | [Blockchain]     | [Blockchain](#Blockchain)  include the blockchains's attributes.
State               | [State]          | The state of the ledger.
Store               | [Store]          | The interface of the sotre.

### Blockchain
Blockchain is a transaction log structured as hash-linked blocks of transactions. 
Peers generate/receive/verify blocks of transactions with other peer, and append the block to the hash chain on the peer’s file system.

Field               | Type              | Description
--------------------|-------------------|----------------------------------------------------------
GenesisBlock        | [Block]          | [Block](#Block) Genesis blocks are defined as deterministic content.
BlockHeight         | [uint32]         | The current block height of this blockchain.
BCEvents            | [events.Event]  | Registered news events ,such as [EventBlockPersistCompleted]
mutex               | [sync.Mutex]    | Determine the preservation of the block with the height +1 atomicity.

### Block
An ordered set of transactions that is cryptographically linked to the preceding block(s).

Field               | Type              | Description
--------------------|-------------------|----------------------------------------------------------
Blockheader         | [Blockheader]     | [Blockheader](#Blockheader) include the block's attributes.
Transactions        | [Transaction]     | List of individual [transactions](#transaction).

##### Genesis Block
The configuration block that initializes a blockchain, and also serves as the first block on a chain.


### Blockheader(Blockdata)
The Hash details of the current Block and the info of prevBlockHash.


Field               | Type              | Description
--------------------|-------------------|----------------------------------------------------------
Version             | uint32            | version of the block which is 0 for now.
Height              | uint32            | Block serial number.
PrevBlockHash       | uint256           | hash value of the previous block.
Timestamp           | uint32            | Time of the block in milliseconds since 00:00:00 UTC Jan 1, 1970.
TransactionsRoot    | Uint256           | Extensible commitment string. See [Block Commitment](#block-commitment).
Nonce               | uint64            | random number.
NextBookkeeper           | Uint160           | NextBookkeeper
Program             | *program.Program  | Program used to validate the block.

### Transaction
Transaction is the base class of all the [Payload](#Payload). Defined with the inputs, outputs and Programs.


Field               | Type              | Description
--------------------|-------------------|----------------------------------------------------------
TxType              | [TransactionType]| For different transaction types with different payload format and transaction process methods.
PayloadVersion      | byte             | PayloadVersion.
Payload             | [Payload]       | Payload.
Nonce               | uint64           | Random number.
Attributes          | []*TxAttribute   | Descirbte the specific attributes of transcation
UTXOInputs          | []*UTXOTxInput   | UTXO module.
BalanceInputs       | []*BalanceTxInput| Balance module.
Outputs             | []*TxOutput      | The Outputs of the transaction.
Program             | []*program.Program | Program used to validate the block.
AssetOutputs        | map[Uint256][]*TxOutput | Outputs asset type.
AssetInputAmount    | map[Uint256]Fixed64 | Inputs map base on Asset.
AssetOutputAmount   | map[Uint256]Fixed64 | Outputs map base on Asset.

#### Payload
Payload is the specific transaction implementtion.

* RegisterAsset

```
RegisterAsset payload be used when register new asset with asset fields:

 "asset name"
 "percision"
 "asset amount"

```
* IssueAsset
* BookkeepingPayload
* SmartContractPayload
* other more payload exentsion.

### Contract
Contract include the program codes with parameters which can be executed on specific evnrioment.
Contract address is the hash of contract program .which be used to control asset or indicate the smart contract address.

Field               | Type              | Description
--------------------|-------------------|----------------------------------------------------------
Code                | []byte            | the contract program code,which will be run on VM or specific environment.
Parameters          | []ContractParameterType| describe the number of contract program parameters and the parameter type
ProgramHash         | Uint160           | The program hash as contract address
OwnerPubkeyHash     | Uint160           | owner's pubkey hash indicate the owner of contract



