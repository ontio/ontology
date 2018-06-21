package utils

import "github.com/ontio/ontology/smartcontract/service/native"

func CheckDirectCall(nativeService *native.NativeService) bool {
	callingAddress := nativeService.ContextRef.CallingContext().ContractAddress
	entryAddress := nativeService.ContextRef.EntryContext().ContractAddress
	if callingAddress == entryAddress {
		return true
	}
	return false
}
