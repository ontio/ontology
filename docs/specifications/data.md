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
Serialization Flags | byte              | See [Block Serialization Flags](#block-serialization-flags).
Version             | varint63          | Block version, equals 1.
Height              | varint63          | Block serial number.
Previous Block ID   | sha3-256          | [Hash](#block-id) of the previous block or all-zero string.
Timestamp           | varint63          | Time of the block in milliseconds since 00:00:00 UTC Jan 1, 1970.
Block Commitment    | Extensible string | Extensible commitment string. See [Block Commitment](#block-commitment).
Block Witness       | Extensible string | Extensible witness string. See [Block Witness](#block-witness).
Transaction Count   | varint31          | Number of transactions that follow.
Transactions        | [Transaction]     | List of individual [transactions](#transaction).

     
### Transaction

### Hash