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
package abi

import "strings"

const (
	NEOVM_PARAM_TYPE_BOOL       = "boolean"
	NEOVM_PARAM_TYPE_STRING     = "string"
	NEOVM_PARAM_TYPE_INTEGER    = "integer"
	NEOVM_PARAM_TYPE_ARRAY      = "array"
	NEOVM_PARAM_TYPE_BYTE_ARRAY = "bytearray"
	NEOVM_PARAM_TYPE_VOID       = "void"
	NEOVM_PARAM_TYPE_ANY        = "any"
)

type NeovmContractAbi struct {
	Address    string                      `json:"hash"`
	EntryPoint string                      `json:"entrypoint"`
	Functions  []*NeovmContractFunctionAbi `json:"functions"`
	Events     []*NeovmContractEventAbi    `json:"events"`
}

func (this *NeovmContractAbi) GetFunc(method string) *NeovmContractFunctionAbi {
	method = strings.ToLower(method)
	for _, funcAbi := range this.Functions {
		if strings.ToLower(funcAbi.Name) == method {
			return funcAbi
		}
	}
	return nil
}

func (this *NeovmContractAbi) GetEvent(evt string) *NeovmContractEventAbi {
	evt = strings.ToLower(evt)
	for _, evtAbi := range this.Events {
		if strings.ToLower(evtAbi.Name) == evt {
			return evtAbi
		}
	}
	return nil
}

type NeovmContractFunctionAbi struct {
	Name       string                    `json:"name"`
	Parameters []*NeovmContractParamsAbi `json:"parameters"`
	ReturnType string                    `json:"returntype"`
}

type NeovmContractParamsAbi struct {
	Name string `json:"name"`
	Type string `json:"type"`
}

type NeovmContractEventAbi struct {
	Name       string                    `json:"name"`
	Parameters []*NeovmContractParamsAbi `json:"parameters"`
	ReturnType string                    `json:"returntype"`
}
