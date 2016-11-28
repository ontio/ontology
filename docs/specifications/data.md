# Onchain DNA Data Model Specification

* [Introduction](#introduction)
* [Definitions](#definitions)
  * [Ledger](#ledger)
 


## Introduction

This document describes the serialization format for the data structures used in the Onchain DNA.

## Definitions

### Block

Field               | Type              | Description
--------------------|-------------------|----------------------------------------------------------
Blockheader         | [Blockheader]     | [Blockheader](#Blockheader) include the block's attributes.
Transactions        | [Transaction]     | List of individual [transactions](#transaction).


### Blockheader


Field               | Type              | Description
--------------------|-------------------|----------------------------------------------------------
Version             | uint32            | version of the block which is 0 for now.
Height              | uint32            | Block serial number.
PrevBlockHash       | uint256           | hash value of the previous block.
Timestamp           | uint32            | Time of the block in milliseconds since 00:00:00 UTC Jan 1, 1970.
TransactionsRoot    | [Hash]            | Extensible commitment string. See [Block Commitment](#block-commitment).
Nonce               | uint64            | random number.
Witness             | [][]byte          | Script used to validate the block.

### Transaction

### Hash






