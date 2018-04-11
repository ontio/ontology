package neovm

import (
	vm "github.com/ontio/ontology/vm/neovm"
)

// get current execute transaction
func GetCodeContainer(service *NeoVmService, engine *vm.ExecutionEngine) error {
	vm.PushData(engine, service.Tx)
	return nil
}

// get current contract address
func GetExecutingAddress(service *NeoVmService, engine *vm.ExecutionEngine) error {
	vm.PushData(engine, service.ContextRef.CurrentContext().ContractAddress[:])
	return nil
}

// get previous call contract address
func GetCallingAddress(service *NeoVmService, engine *vm.ExecutionEngine) error {
	vm.PushData(engine, service.ContextRef.CallingContext().ContractAddress[:])
	return nil
}

// get entry call contract address
func GetEntryAddress(service *NeoVmService, engine *vm.ExecutionEngine) error {
	vm.PushData(engine, service.ContextRef.EntryContext().ContractAddress[:])
	return nil
}

