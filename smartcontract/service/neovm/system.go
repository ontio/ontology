package neovm

import (
	vm "github.com/ontio/ontology/vm/neovm"
	"github.com/ontio/ontology/errors"
)

// get current execute transaction
func GetCodeContainer(service *NeoVmService, engine *vm.ExecutionEngine) error {
	vm.PushData(engine, service.Tx)
	return nil
}

// get current contract address
func GetExecutingAddress(service *NeoVmService, engine *vm.ExecutionEngine) error {
	context := service.ContextRef.CurrentContext(); if context == nil {
		return errors.NewErr("Current context invalid")
	}
	vm.PushData(engine, context.ContractAddress[:])
	return nil
}

// get previous call contract address
func GetCallingAddress(service *NeoVmService, engine *vm.ExecutionEngine) error {
	context := service.ContextRef.CallingContext(); if context == nil {
		return errors.NewErr("Calling context invalid")
	}
	vm.PushData(engine, context.ContractAddress[:])
	return nil
}

// get entry call contract address
func GetEntryAddress(service *NeoVmService, engine *vm.ExecutionEngine) error {
	context := service.ContextRef.EntryContext(); if context == nil {
		return errors.NewErr("Entry context invalid")
	}
	vm.PushData(engine, context.ContractAddress[:])
	return nil
}

