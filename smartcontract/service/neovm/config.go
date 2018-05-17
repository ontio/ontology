/*
 * Copyright (C) 2018 The ontology Authors
 * This file is part of The ontology library.
 *
 * The ontology is free software: you can redistribute it and/or modify
 * it under the terms of the GNU Lesser General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * The ontology is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU Lesser General Public License for more details.
 *
 * You should have received a copy of the GNU Lesser General Public License
 * along with The ontology.  If not, see <http://www.gnu.org/licenses/>.
 */

package neovm

var (
	//Gas Limit
	TRANSACTION_GAS               uint64 = 30000 // Per transaction base cost.
	BLOCKCHAIN_GETHEADER_GAS      uint64 = 100
	BLOCKCHAIN_GETBLOCK_GAS       uint64 = 200
	BLOCKCHAIN_GETTRANSACTION_GAS uint64 = 100
	BLOCKCHAIN_GETCONTRACT_GAS    uint64 = 100
	CONTRACT_CREATE_GAS           uint64 = 10000000
	CONTRACT_MIGRATE_GAS          uint64 = 10000000
	STORAGE_GET_GAS               uint64 = 100
	STORAGE_PUT_GAS               uint64 = 1000
	STORAGE_DELETE_GAS            uint64 = 100
	RUNTIME_CHECKWITNESS_GAS      uint64 = 200
	RUNTIME_CHECKSIG_GAS          uint64 = 200
	APPCALL_GAS                   uint64 = 10
	TAILCALL_GAS                  uint64 = 10
	SHA1_GAS                      uint64 = 10
	SHA256_GAS                    uint64 = 10
	HASH160_GAS                   uint64 = 20
	HASH256_GAS                   uint64 = 20
	OPCODE_GAS                    uint64 = 1

	// API Name
	ATTRIBUTE_GETUSAGE_NAME = "Neo.Attribute.GetUsage"
	ATTRIBUTE_GETDATA_NAME  = "Neo.Attribute.GetData"

	BLOCK_GETTRANSACTIONCOUNT_NAME = "Neo.Block.GetTransactionCount"
	BLOCK_GETTRANSACTIONS_NAME     = "Neo.Block.GetTransactions"
	BLOCK_GETTRANSACTION_NAME      = "Neo.Block.GetTransaction"
	BLOCKCHAIN_GETHEIGHT_NAME      = "Neo.Blockchain.GetHeight"
	BLOCKCHAIN_GETHEADER_NAME      = "Neo.Blockchain.GetHeader"
	BLOCKCHAIN_GETBLOCK_NAME       = "Neo.Blockchain.GetBlock"
	BLOCKCHAIN_GETTRANSACTION_NAME = "Neo.Blockchain.GetTransaction"
	BLOCKCHAIN_GETCONTRACT_NAME    = "Neo.Blockchain.GetContract"

	HEADER_GETINDEX_NAME         = "Neo.Header.GetIndex"
	HEADER_GETHASH_NAME          = "Neo.Header.GetHash"
	HEADER_GETVERSION_NAME       = "Neo.Header.GetVersion"
	HEADER_GETPREVHASH_NAME      = "Neo.Header.GetPrevHash"
	HEADER_GETTIMESTAMP_NAME     = "Neo.Header.GetTimestamp"
	HEADER_GETCONSENSUSDATA_NAME = "Neo.Header.GetConsensusDat"
	HEADER_GETNEXTCONSENSUS_NAME = "Neo.Header.GetNextConsensus"
	HEADER_GETMERKLEROOT_NAME    = "Neo.Header.GetMerkleRoot"

	TRANSACTION_GETHASH_NAME       = "Neo.Transaction.GetHash"
	TRANSACTION_GETTYPE_NAME       = "Neo.Transaction.GetType"
	TRANSACTION_GETATTRIBUTES_NAME = "Neo.Transaction.GetAttributes"

	CONTRACT_CREATE_NAME            = "Neo.Contract.Create"
	CONTRACT_MIGRATE_NAME           = "Neo.Contract.Migrate"
	CONTRACT_GETSTORAGECONTEXT_NAME = "Neo.Contract.GetStorageContext"
	CONTRACT_DESTROY_NAME           = "Neo.Contract.Destroy"
	CONTRACT_GETSCRIPT_NAME         = "Neo.Contract.GetScript"

	STORAGE_GET_NAME        = "Neo.Storage.Get"
	STORAGE_PUT_NAME        = "Neo.Storage.Put"
	STORAGE_DELETE_NAME     = "Neo.Storage.Delete"
	STORAGE_GETCONTEXT_NAME = "Neo.Storage.GetContext"

	RUNTIME_GETTIME_NAME      = "Neo.Runtime.GetTime"
	RUNTIME_CHECKWITNESS_NAME = "Neo.Runtime.CheckWitness"
	RUNTIME_CHECKSIG_NAME     = "Neo.Runtime.CheckSig"
	RUNTIME_NOTIFY_NAME       = "Neo.Runtime.Notify"
	RUNTIME_LOG_NAME          = "Neo.Runtime.Log"

	GETSCRIPTCONTAINER_NAME     = "System.ExecutionEngine.GetScriptContainer"
	GETEXECUTINGSCRIPTHASH_NAME = "System.ExecutionEngine.GetExecutingScriptHash"
	GETCALLINGSCRIPTHASH_NAME   = "System.ExecutionEngine.GetCallingScriptHash"
	GETENTRYSCRIPTHASH_NAME     = "System.ExecutionEngine.GetEntryScriptHash"

	APPCALL_NAME  = "APPCALL"
	TAILCALL_NAME = "TAILCALL"
	SHA1_NAME     = "SHA1"
	SHA256_NAME   = "SHA256"
	HASH160_NAME  = "HASH160"
	HASH256_NAME  = "HASH256"

	GAS_TABLE = map[string]uint64{
		BLOCKCHAIN_GETHEADER_NAME:      BLOCKCHAIN_GETHEADER_GAS,
		BLOCKCHAIN_GETBLOCK_NAME:       BLOCKCHAIN_GETBLOCK_GAS,
		BLOCKCHAIN_GETTRANSACTION_NAME: BLOCKCHAIN_GETTRANSACTION_GAS,
		BLOCKCHAIN_GETCONTRACT_NAME:    BLOCKCHAIN_GETCONTRACT_GAS,
		CONTRACT_CREATE_NAME:           CONTRACT_CREATE_GAS,
		CONTRACT_MIGRATE_NAME:          CONTRACT_MIGRATE_GAS,
		STORAGE_GET_NAME:               STORAGE_GET_GAS,
		STORAGE_PUT_NAME:               STORAGE_PUT_GAS,
		STORAGE_DELETE_NAME:            STORAGE_DELETE_GAS,
		RUNTIME_CHECKWITNESS_NAME:      RUNTIME_CHECKWITNESS_GAS,
		RUNTIME_CHECKSIG_NAME:          RUNTIME_CHECKSIG_GAS,
		APPCALL_NAME:                   APPCALL_GAS,
		TAILCALL_NAME:                  TAILCALL_GAS,
		SHA1_NAME:                      SHA1_GAS,
		SHA256_NAME:                    SHA256_GAS,
		HASH160_NAME:                   HASH160_GAS,
		HASH256_NAME:                   HASH256_GAS,
	}
)
