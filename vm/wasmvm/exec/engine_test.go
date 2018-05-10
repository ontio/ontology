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
	"encoding/binary"
	"io/ioutil"
	"math"
	"testing"

	"github.com/ontio/ontology/common"
)

var service = NewInteropService()

var gasChk = func(uint64) bool {
	return true
}

func TestAdd(t *testing.T) {
	engine := NewExecutionEngine(nil, nil, nil)

	code, err := ioutil.ReadFile("./test_data2/math.wasm")
	if err != nil {
		t.Errorf("TestAdd failed:%s", err.Error())
	}

	method2 := "add"
	input2 := make([]byte, 9)
	input2[0] = byte(len(method2))
	copy(input2[1:len(method2)+1], []byte(method2))
	input2[len(method2)+1] = byte(2) //param count
	input2[len(method2)+2] = byte(1) //param1 length
	input2[len(method2)+3] = byte(1) //param2 length
	input2[len(method2)+4] = byte(5) //param1
	input2[len(method2)+5] = byte(9) //param2

	res2, err := engine.Call(common.Address{}, code, "", input2, 0, gasChk)
	if err != nil {
		t.Errorf("TestAdd failed:%s", err.Error())
	}
	if binary.LittleEndian.Uint32(res2) != uint32(14) {
		t.Error("the result should be 14")
	}

}

func TestSquare(t *testing.T) {
	engine := NewExecutionEngine(nil, nil, nil)

	code, err := ioutil.ReadFile("./test_data2/math.wasm")
	if err != nil {
		t.Errorf("TestAdd err:%s", err.Error())
		return
	}
	method := "square"

	input := make([]byte, 10)
	input[0] = byte(len(method))
	copy(input[1:len(method)+1], []byte(method))
	input[len(method)+1] = byte(1) //param count
	input[len(method)+2] = byte(1) //param1 length
	input[len(method)+3] = byte(5) //param1

	res, err := engine.Call(common.Address{}, code, "", input, 0, gasChk)
	if err != nil {
		t.Errorf("TestAdd err:%s", err.Error())
	}
	if binary.LittleEndian.Uint32(res) != uint32(25) {
		t.Error("the result should be 25")
	}

}

func TestEnvAddTwo(t *testing.T) {

	service.Register("addOne", func(engine *ExecutionEngine) (bool, error) {
		param := engine.vm.envCall.envParams[0]
		engine.vm.ctx = engine.vm.envCall.envPreCtx
		engine.vm.pushUint64(param + 1)
		return true, nil
	})

	engine := NewExecutionEngine(nil, nil, service)

	code, err := ioutil.ReadFile("./test_data2/testenv.wasm")
	if err != nil {
		t.Errorf("TestEnvAddTwo err:%s", err.Error())
	}
	method := "addTwo"

	input := make([]byte, 8)
	input[0] = byte(len(method))
	copy(input[1:len(method)+1], []byte(method))
	input[len(method)+1] = byte(0)

	res, err := engine.Call(common.Address{}, code, "", input, 0, gasChk)
	if err != nil {
		t.Errorf("TestEnvAddTwo err:%s", err.Error())
	}
	if binary.LittleEndian.Uint32(res) != uint32(3) {
		t.Error("the result should be 3")
	}
}

func TestBlockHeight(t *testing.T) {

	service.Register("getBlockHeight", func(engine *ExecutionEngine) (bool, error) {

		engine.vm.ctx = engine.vm.envCall.envPreCtx
		engine.vm.pushUint64(uint64(25535))
		return true, nil
	})

	engine := NewExecutionEngine(nil, nil, service)

	code, err := ioutil.ReadFile("./test_data2/testBlockHeight.wasm")
	if err != nil {
		t.Errorf("TestBlockHeight err:%s", err.Error())
	}
	method := "blockHeight"

	input := make([]byte, 13)
	input[0] = byte(len(method))
	copy(input[1:len(method)+1], []byte(method))
	input[len(method)+1] = byte(0)

	res, err := engine.Call(common.Address{}, code, "", input, 0, gasChk)
	if err != nil {
		t.Errorf("TestBlockHeight err:%s", err.Error())
	}
	if binary.LittleEndian.Uint32(res) != uint32(25535) {
		t.Error("the result should be 25535")
	}
}

func TestMem(t *testing.T) {

	service.Register("getString", func(engine *ExecutionEngine) (bool, error) {

		mem := engine.vm.memory.Memory
		param := engine.vm.envCall.envParams
		start := int(param[0])
		length := int(param[1])
		str := string(mem[start : start+length])
		engine.vm.ctx = engine.vm.envCall.envPreCtx
		engine.vm.pushUint64(uint64(len(str)))
		return true, nil

	})

	engine := NewExecutionEngine(nil, nil, service)
	//test
	code, err := ioutil.ReadFile("./test_data2/TestMemory.wasm")
	if err != nil {
		t.Errorf("TestMem err:%s", err.Error())
	}
	method := "getStrLen"

	input := make([]byte, 11)
	input[0] = byte(len(method))
	copy(input[1:len(method)+1], []byte(method))
	input[len(method)+1] = byte(0)

	res, err := engine.Call(common.Address{}, code, "", input, 0, gasChk)
	if err != nil {
		t.Errorf("TestMem err:%s", err.Error())
	}
	if binary.LittleEndian.Uint32(res) != uint32(11) {
		t.Error("the result should be 11")
	}
}

func TestGlobal(t *testing.T) {
	engine := NewExecutionEngine(nil, nil, nil)
	//test
	code, err := ioutil.ReadFile("./test_data2/str.wasm")
	if err != nil {
		t.Errorf("TestGlobal err:%s", err.Error())
	}
	method := "_getStr"

	input := make([]byte, 9)
	input[0] = byte(len(method))
	copy(input[1:len(method)+1], []byte(method))
	input[len(method)+1] = byte(0)

	_, err = engine.Call(common.Address{}, code, "", input, 0, gasChk)
	if err != nil {
		t.Errorf("TestGlobal err:%s", err.Error())
	}

}

func TestIf(t *testing.T) {
	engine := NewExecutionEngine(nil, nil, nil)
	//test
	code, err := ioutil.ReadFile("./test_data2/ifTest.wasm")
	if err != nil {
		t.Errorf("TestIf err:%s", err.Error())
	}
	method := "testif"
	p := 5

	input := make([]byte, 100)
	input[0] = byte(len(method))
	copy(input[1:len(method)+1], []byte(method))
	input[len(method)+1] = byte(1)
	input[len(method)+2] = byte(1)
	input[len(method)+3] = byte(p)

	res, err := engine.Call(common.Address{}, code, "", input, 0, gasChk)
	if err != nil {
		t.Errorf("TestIf err:%s", err.Error())
	}

	var shouldR int
	if p < 5 {
		shouldR = 10 + p
	} else {
		shouldR = 20 + p
	}

	if binary.LittleEndian.Uint32(res) != uint32(shouldR) {
		t.Fatal("result should be", shouldR)
	}

}

func TestLoop(t *testing.T) {
	engine := NewExecutionEngine(nil, nil, nil)
	//test
	code, err := ioutil.ReadFile("./test_data2/ifTest.wasm")
	if err != nil {
		t.Errorf("TestLoop err:%s", err.Error())
	}
	method := "testfor"
	p := 10

	input := make([]byte, 100)
	input[0] = byte(len(method))
	copy(input[1:len(method)+1], []byte(method))
	input[len(method)+1] = byte(1)
	input[len(method)+2] = byte(1)
	input[len(method)+3] = byte(p)

	res, err := engine.Call(common.Address{}, code, "", input, 0, gasChk)
	if err != nil {
		t.Errorf("TestLoop err:%s", err.Error())
	}

	if binary.LittleEndian.Uint32(res) != uint32(2*(p+1)) {
		t.Fatal("result should be", 2*(p+1))
	}

}

func TestWhileLoop(t *testing.T) {
	engine := NewExecutionEngine(nil, nil, nil)
	//test
	code, err := ioutil.ReadFile("./test_data2/ifTest.wasm")
	if err != nil {
		t.Errorf("TestWhileLoop err:%s", err.Error())
	}
	method := "testwhile"
	p := 10

	input := make([]byte, 100)
	input[0] = byte(len(method))
	copy(input[1:len(method)+1], []byte(method))
	input[len(method)+1] = byte(1)
	input[len(method)+2] = byte(1)
	input[len(method)+3] = byte(p)

	res, err := engine.Call(common.Address{}, code, "", input, 0, gasChk)
	if err != nil {
		t.Errorf("TestWhileLoop err:%s", err.Error())
	}

	if binary.LittleEndian.Uint32(res) != uint32(p+10) {
		t.Fatal("result should be", p+10)
	}

}

func TestIfII(t *testing.T) {

	s := "b456c4862902525e17ace6a2607f0806f51df0a98c3629c27f00efcf87ee8784"
	u := binary.LittleEndian.Uint64([]byte(s))

	b := make([]byte, 64)
	binary.LittleEndian.PutUint64(b, u)

	engine := NewExecutionEngine(nil, nil, nil)
	//test
	code, err := ioutil.ReadFile("./test_data2/ifTest.wasm")
	if err != nil {
		t.Errorf("TestIfII err:%s", err.Error())
	}
	method := "testifII"
	p := 10

	input := make([]byte, 100)
	input[0] = byte(len(method))
	copy(input[1:len(method)+1], []byte(method))
	input[len(method)+1] = byte(1)
	input[len(method)+2] = byte(1)
	input[len(method)+3] = byte(p)

	res, err := engine.Call(common.Address{}, code, "", input, 0, gasChk)
	if err != nil {
		t.Errorf("TestIfII err:%s", err.Error())
	}

	if binary.LittleEndian.Uint32(res) != uint32(60) {
		t.Fatal("result should be", 60)
	}

}

func TestStrings(t *testing.T) {
	engine := NewExecutionEngine(nil, nil, nil)
	//test
	code, err := ioutil.ReadFile("./test_data2/strings.wasm")
	if err != nil {
		t.Errorf("TestStrings err:%s", err.Error())
	}

	input := make([]interface{}, 2)
	input[0] = "getStringlen"
	input[1] = "abcdefghijklmnopqrstuvwxyz"
	//input[2] = 3

	res, err := engine.CallInf(common.Address{}, code, input, nil, gasChk)
	if err != nil {
		t.Errorf("TestStrings err:%s", err.Error())
	}

	if binary.LittleEndian.Uint32(res) != 26 {
		t.Fatal("the res should be 26")
	}

}

func TestIntArraySum(t *testing.T) {
	engine := NewExecutionEngine(nil, nil, nil)
	//test
	code, err := ioutil.ReadFile("./test_data2/intarray.wasm")
	if err != nil {
		t.Errorf("TestIntArraySum err:%s", err.Error())
		return
	}

	input := make([]interface{}, 3)
	input[0] = "_sum"
	input[1] = []int{1, 2, 3, 4}
	input[2] = 4

	res, err := engine.CallInf(common.Address{}, code, input, nil, gasChk)
	if err != nil {
		t.Errorf("TestIntArraySum err:%s", err.Error())
	}

	if binary.LittleEndian.Uint32(res) != 10 {
		t.Fatal("the res should be 10")
	}

}

func TestSimplestruct(t *testing.T) {
	engine := NewExecutionEngine(nil, nil, nil)
	//test
	code, err := ioutil.ReadFile("./test_data2/simplestruct.wasm")
	if err != nil {
		t.Errorf("TestSimplestruct err:%s", err.Error())

	}

	type student struct {
		name string
		math int
		eng  int64
	}

	s := student{name: "jack", math: 90, eng: 95}

	input := make([]interface{}, 2)
	input[0] = "getSum"
	input[1] = s

	res, err := engine.CallInf(common.Address{}, code, input, nil, gasChk)
	if err != nil {
		t.Errorf("TestSimplestruct err:%s", err.Error())
	}
	if binary.LittleEndian.Uint32(res) != 185 {
		t.Fatal("the res should be 185")
	}

}

func TestSimplestruct2(t *testing.T) {
	engine := NewExecutionEngine(nil, nil, nil)
	//test
	code, err := ioutil.ReadFile("./test_data2/simplestruct.wasm")
	if err != nil {
		t.Errorf("TestSimplestruct2 err:%s", err.Error())

	}

	type student struct {
		name string
		math int
		eng  int64
	}

	s := student{name: "jack", math: 90, eng: 95}

	input := make([]interface{}, 2)
	input[0] = "getName"
	input[1] = s

	res, err := engine.CallInf(common.Address{}, code, input, nil, gasChk)
	if err != nil {
		t.Errorf("TestSimplestruct2 err:%s", err.Error())
	}
	idx := int(binary.LittleEndian.Uint32(res))
	var length int
	for i := idx; engine.vm.memory.Memory[i] != 0; i++ {
		length += 1
	}
	if string(engine.vm.memory.Memory[idx:idx+length]) != "jack" {
		t.Fatal("the res should be jack")
	}

}

func TestFloatSum(t *testing.T) {
	engine := NewExecutionEngine(nil, nil, nil)
	//test
	code, err := ioutil.ReadFile("./test_data2/float.wasm")
	if err != nil {
		t.Errorf("TestFloatSum err:%s", err.Error())
	}

	input := make([]interface{}, 3)
	input[0] = "sum"
	input[1] = float32(1.1)
	input[2] = float32(0.5)

	res, err := engine.CallInf(common.Address{}, code, input, nil, gasChk)
	if err != nil {
		t.Errorf("TestFloatSum err:%s", err.Error())
	}

	if math.Float32frombits(binary.LittleEndian.Uint32(res)) != 1.6 {
		t.Fatal("the res should be  1.6 ")
	}

}
func TestDoubleSum(t *testing.T) {
	engine := NewExecutionEngine(nil, nil, nil)
	//test
	code, err := ioutil.ReadFile("./test_data2/float.wasm")
	if err != nil {
		t.Errorf("TestDoubleSum err:%s", err.Error())

	}

	input := make([]interface{}, 3)
	input[0] = "sumDouble"
	input[1] = 1.1
	input[2] = 0.5

	res, err := engine.CallInf(common.Address{}, code, input, nil, gasChk)
	if err != nil {
		t.Errorf("TestDoubleSum err:%s", err.Error())
	}

	if math.Float64frombits(binary.LittleEndian.Uint64(res)) != 1.6 {
		t.Fatal("the res should be  1.6 ")
	}
}

func TestCalloc(t *testing.T) {
	engine := NewExecutionEngine(nil, nil, nil)
	//test
	code, err := ioutil.ReadFile("./test_data2/calloc.wasm")
	if err != nil {
		t.Errorf("TestCalloc err:%s", err.Error())

	}

	input := make([]interface{}, 1)
	input[0] = "retArray"

	_, err = engine.CallInf(common.Address{}, code, input, nil, gasChk)
	if err != nil {
		t.Errorf("TestCalloc err:%s", err.Error())
	}

}

func TestMalloc(t *testing.T) {
	engine := NewExecutionEngine(nil, nil, nil)
	//test
	code, err := ioutil.ReadFile("./test_data2/malloc.wasm")
	if err != nil {
		t.Errorf("TestMalloc err:%s", err.Error())
	}

	input := make([]interface{}, 4)
	input[0] = "initStu"
	input[1] = 100
	input[2] = 90
	input[3] = 85

	_, err = engine.CallInf(common.Address{}, code, input, nil, gasChk)
	if err != nil {
		t.Errorf("TestMalloc err:%s", err.Error())
	}

}

//use 'arrayLen' instead of  'sizeof'
func TestArraylen(t *testing.T) {
	engine := NewExecutionEngine(nil, nil, nil)
	//test
	code, err := ioutil.ReadFile("./test_data2/arraylen.wasm")
	if err != nil {
		t.Errorf("TestArraylen err:%s", err.Error())

	}

	input := make([]interface{}, 3)
	input[0] = "combine"
	input[1] = []int{1, 2, 3, 4, 5}
	input[2] = []int{6, 7, 8, 9, 10, 11}

	res, err := engine.CallInf(common.Address{}, code, input, nil, gasChk)
	if err != nil {
		t.Errorf("TestArraylen err:%s", err.Error())
	}

	offset := binary.LittleEndian.Uint32(res)
	bytes, _ := engine.vm.memory.GetPointerMemory(uint64(offset))
	cnt := len(bytes) / 4
	resarray := make([]int, cnt)
	for i := 0; i < cnt; i++ {
		resarray[i] = int(binary.LittleEndian.Uint32(bytes[i*4 : (i+1)*4]))
	}
	if len(resarray) != 11 {
		t.Fatal("TestArraylen result should be 11")
	}
}

func TestAddress(t *testing.T) {
	engine := NewExecutionEngine(nil, nil, nil)
	//test
	code, err := ioutil.ReadFile("./test_data2/testGetAddress.wasm")
	if err != nil {
		t.Errorf("TestAddress err:%s", err.Error())
	}

	input := make([]interface{}, 1)
	input[0] = "_getaddr"

	_, err = engine.CallInf(common.Address{}, code, input, nil, gasChk)
	if err != nil {
		t.Errorf("TestAddress err:%s", err.Error())
	}
}

func TestString(t *testing.T) {
	engine := NewExecutionEngine(nil, nil, nil)
	//test
	code, err := ioutil.ReadFile("./test_data2/stringtest.wasm")
	if err != nil {
		t.Errorf("TestString err:%s", err.Error())

	}

	input := make([]interface{}, 2)
	input[0] = "greeting"
	input[1] = "may the force be with you" //code ,address

	res, err := engine.CallInf(common.Address{}, code, input, nil, gasChk)
	if err != nil {
		t.Errorf("TestString err:%s", err.Error())
	}

	offset := uint64(binary.LittleEndian.Uint32(res))
	length := engine.vm.memory.MemPoints[offset].Length

	if input[1] != string(engine.vm.memory.Memory[offset:int(offset)+int(length)]) {
		t.Fatal("the return should be :" + input[1].(string))
	}

}
