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
	"fmt"
	"github.com/ontio/ontology/vm/neovm/errors"
	"github.com/ontio/ontology/vm/neovm/types"
	"reflect"
)

func NewExecutionEngine() *ExecutionEngine {
	var engine ExecutionEngine
	engine.EvaluationStack = NewRandAccessStack()
	engine.AltStack = NewRandAccessStack()
	engine.State = BREAK
	engine.OpCode = 0
	return &engine
}

type ExecutionEngine struct {
	EvaluationStack *RandomAccessStack
	AltStack        *RandomAccessStack
	State           VMState
	Contexts        []*ExecutionContext
	Context         *ExecutionContext
	OpCode          OpCode
	OpExec          OpExec
}

func (this *ExecutionEngine) CurrentContext() *ExecutionContext {
	return this.Contexts[len(this.Contexts)-1]
}

func (this *ExecutionEngine) PopContext() {
	if len(this.Contexts) != 0 {
		this.Contexts = this.Contexts[:len(this.Contexts)-1]
	}
	if len(this.Contexts) != 0 {
		this.Context = this.CurrentContext()
	}
}

func (this *ExecutionEngine) PushContext(context *ExecutionContext) {
	this.Contexts = append(this.Contexts, context)
	this.Context = this.CurrentContext()
}

func (this *ExecutionEngine) Execute() error {
	this.State = this.State & (^BREAK)
	for {
		if this.State == FAULT || this.State == HALT || this.State == BREAK {
			break
		}
		err := this.StepInto()
		if err != nil {
			return err
		}
	}
	return nil
}

func (this *ExecutionEngine) ExecuteCode() error {
	code, err := this.Context.OpReader.ReadByte()
	if err != nil {
		this.State = FAULT
		return err
	}
	this.OpCode = OpCode(code)
	return nil
}

func (this *ExecutionEngine) ValidateOp() error {
	opExec := OpExecList[this.OpCode]
	if opExec.Name == "" {
		return errors.ERR_NOT_SUPPORT_OPCODE
	}
	this.OpExec = opExec
	return nil
}

func (this *ExecutionEngine) StepInto() error {
	state, err := this.ExecuteOp()
	this.State = state
	if err != nil {
		return err
	}
	return nil
}

func (this *ExecutionEngine) ExecuteOp() (VMState, error) {
	if this.OpCode >= PUSHBYTES1 && this.OpCode <= PUSHBYTES75 {
		bs, err := this.Context.OpReader.ReadBytes(int(this.OpCode))
		if err != nil {
			return FAULT, err
		}
		PushData(this, bs)
		return NONE, nil
	}

	fmt.Println("op:", this.OpExec.Name)
	s := this.EvaluationStack.Count()
	for i := 0; i < s; i++ {
		item := this.EvaluationStack.Peek(i)
		fmt.Print("type:", reflect.TypeOf(item))
		fmt.Print(" ")
		switch v := item.(type) {
		case *types.Integer:
			s, _ := v.GetBigInteger()
			fmt.Printf("value:%v", s)
		case *types.Boolean:
			s, _ := v.GetBoolean()
			fmt.Printf("value:%v", s)
		case *types.ByteArray:
			s, _ := v.GetByteArray()
			fmt.Printf("value:%v", s)
		case *types.Interop:
			s, _ := v.GetInterface()
			fmt.Printf("value:%v", s)
		case *types.Array:
			s, _ := v.GetArray()
			fmt.Printf("value:%v", s)
		case *types.Map:
			s, _ := v.GetMap()
			for k, v := range s {
				fmt.Printf("key:%v value:%v", k, v)
			}
		}
		fmt.Print(" ")
	}
	fmt.Println()

	if this.OpExec.Validator != nil {
		if err := this.OpExec.Validator(this); err != nil {
			return FAULT, err
		}
	}
	return this.OpExec.Exec(this)
}
