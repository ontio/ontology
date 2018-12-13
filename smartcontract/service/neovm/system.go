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

import (
	"github.com/ontio/ontology/errors"
	vm "github.com/ontio/ontology/vm/neovm"
)

// GetCodeContainer push current transaction to vm stack
func GetCodeContainer(service *NeoVmService, engine *vm.Executor) error {
	return engine.EvalStack.PushAsInteropValue(service.Tx)
}

// GetExecutingAddress push current context to vm stack
func GetExecutingAddress(service *NeoVmService, engine *vm.Executor) error {
	context := service.ContextRef.CurrentContext()
	if context == nil {
		return errors.NewErr("Current context invalid")
	}
	return engine.EvalStack.PushBytes(context.ContractAddress[:])
}

// GetExecutingAddress push previous context to vm stack
func GetCallingAddress(service *NeoVmService, engine *vm.Executor) error {
	context := service.ContextRef.CallingContext()
	if context == nil {
		return errors.NewErr("Calling context invalid")
	}
	return engine.EvalStack.PushBytes(context.ContractAddress[:])
}

// GetExecutingAddress push entry call context to vm stack
func GetEntryAddress(service *NeoVmService, engine *vm.Executor) error {
	context := service.ContextRef.EntryContext()
	if context == nil {
		return errors.NewErr("Entry context invalid")
	}
	return engine.EvalStack.PushBytes(context.ContractAddress[:])
}
