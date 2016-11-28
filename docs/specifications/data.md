# Onchain DNA Data Model Specification

* [Introduction](#introduction)
* [Definitions](#definitions)
  * [Ledger](#ledger)
 


## Introduction

This document describes the serialization format for the data structures used in the Onchain DNA.

## Definitions

### Block

      |Size|Field|DataType|Description|
      |---|---|---|---|
      ||Blockheader|[Blockheader]|[Blockheader](#Blockheader) include the block's attributes.|
      ||Transactions|[Transaction]|List of individual [transactions](#transaction).|

### Blockheader

   |Size|Field|DataType   |Description|
   |--- |---|---        |---|
   | |Version        |uint32     |version of the block which is 0 for now|
   | |Height     |uint32     |height of block|
   | |PrevBlockHash  |uint256 [Hash]   |hash value of the previous block|
   | |Timestamp  |uint32     |time stamp|
   | |TransactionsRoot     |[Hash]    |root hash of a transaction list|
   | |Nonce      |uint64     |random number|
   | |Witness     |[][]byte     |Script used to validate the block|
     
### Transaction

### Hash