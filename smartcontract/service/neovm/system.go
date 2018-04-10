package neovm

import (
	vm "github.com/ontio/ontology/vm/neovm"
)

func GetCodeContainer(service *NeoVmService, engine *vm.ExecutionEngine) error {
	vm.PushData(engine, service.Tx)
	return nil
}

func GetExecutingAddress(service *NeoVmService, engine *vm.ExecutionEngine) error {
	vm.PushData(engine, service.ContextRef.CurrentContext().ContractAddress[:])
	return nil
}

func GetCallingAddress(service *NeoVmService, engine *vm.ExecutionEngine) error {
	vm.PushData(engine, service.ContextRef.CallingContext().ContractAddress[:])
	return nil
}

func GetEntryAddress(service *NeoVmService, engine *vm.ExecutionEngine) error {
	vm.PushData(engine, service.ContextRef.EntryContext().ContractAddress[:])
	return nil
}

