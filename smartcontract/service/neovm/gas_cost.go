package neovm

import (
	vm "github.com/ontio/ontology/vm/neovm"
)

func StoreGasCost(engine *vm.ExecutionEngine) uint64 {
	key := vm.PeekNByteArray(0, engine)
	value := vm.PeekNByteArray(1, engine)
	return uint64(((len(key)+len(value)-1)/1024 + 1)) * GAS_TABLE[STORAGE_PUT_NAME]
}

func GasPrice(engine *vm.ExecutionEngine, name string) uint64 {
	switch name {
	case STORAGE_PUT_NAME:
		return StoreGasCost(engine)
	default:
		if value, ok := GAS_TABLE[name]; ok {
			return value
		}
		return OPCODE_GAS
	}
}
