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

import "sync"

var (
	//Gas Limit
	MIN_TRANSACTION_GAS           uint64 = 20000 // Per transaction base cost.
	BLOCKCHAIN_GETHEADER_GAS      uint64 = 100
	BLOCKCHAIN_GETBLOCK_GAS       uint64 = 200
	BLOCKCHAIN_GETTRANSACTION_GAS uint64 = 100
	BLOCKCHAIN_GETCONTRACT_GAS    uint64 = 100
	CONTRACT_CREATE_GAS           uint64 = 20000000
	CONTRACT_MIGRATE_GAS          uint64 = 20000000
	UINT_DEPLOY_CODE_LEN_GAS      uint64 = 200000
	UINT_INVOKE_CODE_LEN_GAS      uint64 = 20000
	NATIVE_INVOKE_GAS             uint64 = 1000
	STORAGE_GET_GAS               uint64 = 200
	STORAGE_PUT_GAS               uint64 = 4000
	STORAGE_DELETE_GAS            uint64 = 100
	RUNTIME_CHECKWITNESS_GAS      uint64 = 200
	RUNTIME_VERIFYMUTISIG_GAS     uint64 = 400
	RUNTIME_ADDRESSTOBASE58_GAS   uint64 = 40
	RUNTIME_BASE58TOADDRESS_GAS   uint64 = 30
	APPCALL_GAS                   uint64 = 10
	TAILCALL_GAS                  uint64 = 10
	SHA1_GAS                      uint64 = 10
	SHA256_GAS                    uint64 = 10
	HASH160_GAS                   uint64 = 20
	HASH256_GAS                   uint64 = 20
	OPCODE_GAS                    uint64 = 1

	PER_UNIT_CODE_LEN    int = 1024
	METHOD_LENGTH_LIMIT  int = 1024
	DUPLICATE_STACK_SIZE int = 1024 * 2
	VM_STEP_LIMIT        int = 400000

	// API Name
	ATTRIBUTE_GETUSAGE_NAME = "Ontology.Attribute.GetUsage"
	ATTRIBUTE_GETDATA_NAME  = "Ontology.Attribute.GetData"

	BLOCK_GETTRANSACTIONCOUNT_NAME       = "System.Block.GetTransactionCount"
	BLOCK_GETTRANSACTIONS_NAME           = "System.Block.GetTransactions"
	BLOCK_GETTRANSACTION_NAME            = "System.Block.GetTransaction"
	BLOCKCHAIN_GETHEIGHT_NAME            = "System.Blockchain.GetHeight"
	BLOCKCHAIN_GETHEADER_NAME            = "System.Blockchain.GetHeader"
	BLOCKCHAIN_GETBLOCK_NAME             = "System.Blockchain.GetBlock"
	BLOCKCHAIN_GETTRANSACTION_NAME       = "System.Blockchain.GetTransaction"
	BLOCKCHAIN_GETCONTRACT_NAME          = "System.Blockchain.GetContract"
	BLOCKCHAIN_GETTRANSACTIONHEIGHT_NAME = "System.Blockchain.GetTransactionHeight"

	HEADER_GETINDEX_NAME         = "System.Header.GetIndex"
	HEADER_GETHASH_NAME          = "System.Header.GetHash"
	HEADER_GETVERSION_NAME       = "Ontology.Header.GetVersion"
	HEADER_GETPREVHASH_NAME      = "System.Header.GetPrevHash"
	HEADER_GETTIMESTAMP_NAME     = "System.Header.GetTimestamp"
	HEADER_GETCONSENSUSDATA_NAME = "Ontology.Header.GetConsensusData"
	HEADER_GETNEXTCONSENSUS_NAME = "Ontology.Header.GetNextConsensus"
	HEADER_GETMERKLEROOT_NAME    = "Ontology.Header.GetMerkleRoot"

	TRANSACTION_GETHASH_NAME       = "System.Transaction.GetHash"
	TRANSACTION_GETTYPE_NAME       = "Ontology.Transaction.GetType"
	TRANSACTION_GETATTRIBUTES_NAME = "Ontology.Transaction.GetAttributes"

	CONTRACT_CREATE_NAME            = "Ontology.Contract.Create"
	CONTRACT_MIGRATE_NAME           = "Ontology.Contract.Migrate"
	CONTRACT_GETSTORAGECONTEXT_NAME = "System.Contract.GetStorageContext"
	CONTRACT_DESTROY_NAME           = "System.Contract.Destroy"
	CONTRACT_GETSCRIPT_NAME         = "Ontology.Contract.GetScript"

	STORAGE_GET_NAME                = "System.Storage.Get"
	STORAGE_PUT_NAME                = "System.Storage.Put"
	STORAGE_DELETE_NAME             = "System.Storage.Delete"
	STORAGE_GETCONTEXT_NAME         = "System.Storage.GetContext"
	STORAGE_GETREADONLYCONTEXT_NAME = "System.Storage.GetReadOnlyContext"

	STORAGECONTEXT_ASREADONLY_NAME = "System.StorageContext.AsReadOnly"

	RUNTIME_GETTIME_NAME             = "System.Runtime.GetTime"
	RUNTIME_CHECKWITNESS_NAME        = "System.Runtime.CheckWitness"
	RUNTIME_NOTIFY_NAME              = "System.Runtime.Notify"
	RUNTIME_LOG_NAME                 = "System.Runtime.Log"
	RUNTIME_GETTRIGGER_NAME          = "System.Runtime.GetTrigger"
	RUNTIME_SERIALIZE_NAME           = "System.Runtime.Serialize"
	RUNTIME_DESERIALIZE_NAME         = "System.Runtime.Deserialize"
	RUNTIME_BASE58TOADDRESS_NAME     = "Ontology.Runtime.Base58ToAddress"
	RUNTIME_ADDRESSTOBASE58_NAME     = "Ontology.Runtime.AddressToBase58"
	RUNTIME_GETCURRENTBLOCKHASH_NAME = "Ontology.Runtime.GetCurrentBlockHash"
	RUNTIME_VERIFYMUTISIG_NAME       = "Ontology.Runtime.VerifyMutiSig"

	NATIVE_INVOKE_NAME = "Ontology.Native.Invoke"

	GETSCRIPTCONTAINER_NAME     = "System.ExecutionEngine.GetScriptContainer"
	GETEXECUTINGSCRIPTHASH_NAME = "System.ExecutionEngine.GetExecutingScriptHash"
	GETCALLINGSCRIPTHASH_NAME   = "System.ExecutionEngine.GetCallingScriptHash"
	GETENTRYSCRIPTHASH_NAME     = "System.ExecutionEngine.GetEntryScriptHash"

	APPCALL_NAME              = "APPCALL"
	TAILCALL_NAME             = "TAILCALL"
	SHA1_NAME                 = "SHA1"
	SHA256_NAME               = "SHA256"
	HASH160_NAME              = "HASH160"
	HASH256_NAME              = "HASH256"
	UINT_DEPLOY_CODE_LEN_NAME = "Deploy.Code.Gas"
	UINT_INVOKE_CODE_LEN_NAME = "Invoke.Code.Gas"

	GAS_TABLE = initGAS_TABLE()

	GAS_TABLE_KEYS = []string{
		BLOCKCHAIN_GETHEADER_NAME,
		BLOCKCHAIN_GETBLOCK_NAME,
		BLOCKCHAIN_GETTRANSACTION_NAME,
		BLOCKCHAIN_GETCONTRACT_NAME,
		CONTRACT_CREATE_NAME,
		CONTRACT_MIGRATE_NAME,
		STORAGE_GET_NAME,
		STORAGE_PUT_NAME,
		STORAGE_DELETE_NAME,
		RUNTIME_CHECKWITNESS_NAME,
		NATIVE_INVOKE_NAME,
		APPCALL_NAME,
		TAILCALL_NAME,
		SHA1_NAME,
		SHA256_NAME,
		HASH160_NAME,
		HASH256_NAME,
		UINT_DEPLOY_CODE_LEN_NAME,
		UINT_INVOKE_CODE_LEN_NAME,
	}

	INIT_GAS_TABLE = map[string]uint64{
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
		NATIVE_INVOKE_NAME:             NATIVE_INVOKE_GAS,
		APPCALL_NAME:                   APPCALL_GAS,
		TAILCALL_NAME:                  TAILCALL_GAS,
		SHA1_NAME:                      SHA1_GAS,
		SHA256_NAME:                    SHA256_GAS,
		HASH160_NAME:                   HASH160_GAS,
		HASH256_NAME:                   HASH256_GAS,
		UINT_DEPLOY_CODE_LEN_NAME:      UINT_DEPLOY_CODE_LEN_GAS,
		UINT_INVOKE_CODE_LEN_NAME:      UINT_INVOKE_CODE_LEN_GAS,
	}
)

func initGAS_TABLE() *sync.Map {
	m := sync.Map{}
	m.Store(BLOCKCHAIN_GETHEADER_NAME, BLOCKCHAIN_GETHEADER_GAS)
	m.Store(BLOCKCHAIN_GETBLOCK_NAME, BLOCKCHAIN_GETBLOCK_GAS)
	m.Store(BLOCKCHAIN_GETTRANSACTION_NAME, BLOCKCHAIN_GETTRANSACTION_GAS)
	m.Store(BLOCKCHAIN_GETCONTRACT_NAME, BLOCKCHAIN_GETCONTRACT_GAS)
	m.Store(CONTRACT_CREATE_NAME, CONTRACT_CREATE_GAS)
	m.Store(CONTRACT_MIGRATE_NAME, CONTRACT_MIGRATE_GAS)
	m.Store(STORAGE_GET_NAME, STORAGE_GET_GAS)
	m.Store(STORAGE_PUT_NAME, STORAGE_PUT_GAS)
	m.Store(STORAGE_DELETE_NAME, STORAGE_DELETE_GAS)
	m.Store(RUNTIME_CHECKWITNESS_NAME, RUNTIME_CHECKWITNESS_GAS)
	m.Store(NATIVE_INVOKE_NAME, NATIVE_INVOKE_GAS)
	m.Store(APPCALL_NAME, APPCALL_GAS)
	m.Store(TAILCALL_NAME, TAILCALL_GAS)
	m.Store(SHA1_NAME, SHA1_GAS)
	m.Store(SHA256_NAME, SHA256_GAS)
	m.Store(HASH160_NAME, HASH160_GAS)
	m.Store(HASH256_NAME, HASH256_GAS)
	m.Store(UINT_DEPLOY_CODE_LEN_NAME, UINT_DEPLOY_CODE_LEN_GAS)
	m.Store(UINT_INVOKE_CODE_LEN_NAME, UINT_INVOKE_CODE_LEN_GAS)

	m.Store(RUNTIME_BASE58TOADDRESS_NAME, RUNTIME_BASE58TOADDRESS_GAS)
	m.Store(RUNTIME_ADDRESSTOBASE58_NAME, RUNTIME_ADDRESSTOBASE58_GAS)

	m.Store(RUNTIME_VERIFYMUTISIG_NAME, RUNTIME_VERIFYMUTISIG_GAS)

	return &m
}
