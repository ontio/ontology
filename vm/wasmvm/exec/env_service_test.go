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
package exec

import (
	"errors"
	"testing"
)

func TestNewInteropService(t *testing.T) {

	service := NewInteropService()
	if service == nil {
		t.Error("NewInteropService should not return nil")
	}

	if len(service.serviceMap) == 0 {
		t.Error("NewInteropService servicemap should not be empty")
	}
}

func TestInteropService_Register(t *testing.T) {
	service := InteropService{make(map[string]func(*ExecutionEngine) (bool, error))}

	ret := service.Register("testFuncTrue", func(e *ExecutionEngine) (bool, error) { return true, nil })
	if !ret {
		t.Error("Register should return true!")
	}
	ret = service.Register("testFuncTrue", func(e *ExecutionEngine) (bool, error) { return true, nil })
	if ret {
		t.Error("Register should return false while put same keyname!")
	}

}

func TestInteropService_Exists(t *testing.T) {
	service := InteropService{make(map[string]func(*ExecutionEngine) (bool, error))}

	service.Register("testFuncTrue", func(e *ExecutionEngine) (bool, error) { return true, nil })
	exists := service.Exists("testFuncTrue")
	if !exists {
		t.Error("key should exists in servicemap!")
	}
	exists = service.Exists("testFuncfalse")
	if exists {
		t.Error("key should not exists in servicemap!")
	}
}

func TestInteropService_GetServiceMap(t *testing.T) {
	service := InteropService{make(map[string]func(*ExecutionEngine) (bool, error))}

	service.Register("testFuncTrue", func(e *ExecutionEngine) (bool, error) { return true, nil })

	smap := service.GetServiceMap()
	_, ok := smap["testFuncTrue"]
	if !ok {
		t.Error("'testFuncTrue' should exist in servicemap")
	}
}

func TestInteropService_MergeMap(t *testing.T) {
	service := InteropService{make(map[string]func(*ExecutionEngine) (bool, error))}

	service.Register("testFuncTrue", func(e *ExecutionEngine) (bool, error) { return true, nil })

	smap := make(map[string]func(*ExecutionEngine) (bool, error))
	smap["testFuncFalse"] = func(e *ExecutionEngine) (bool, error) { return false, errors.New("some error happened") }

	res := service.MergeMap(smap)
	if !res {
		t.Error("mergeMap should return true!")
	}

	if !service.Exists("testFuncFalse") {
		t.Error("'testFuncFalse' should exist in merged map")
	}
}

func TestInteropService_Invoke(t *testing.T) {
	service := InteropService{make(map[string]func(*ExecutionEngine) (bool, error))}

	service.Register("testFuncTrue", func(e *ExecutionEngine) (bool, error) { return true, nil })
	//service.Register("testFuncFalse",func(e *ExecutionEngine)(bool,error){return false,errors.New("some error happened")})

	_, err := service.Invoke("testFuncTrue", &ExecutionEngine{})
	if err != nil {
		t.Error("invoke testFuncTrue should return true and no error!")
	}
	_, err = service.Invoke("notexistfunc", &ExecutionEngine{})
	if err == nil {
		t.Error("invoke 'notexistfunc' shou")
	}

}
