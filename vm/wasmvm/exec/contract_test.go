package exec

import (
	"testing"
	"io/ioutil"
	"fmt"
	"github.com/Ontology/common"
	"encoding/json"
	"encoding/binary"
	"bytes"
	"github.com/Ontology/common/serialization"
)

func TestContract1(t *testing.T){
	engine := NewExecutionEngine(nil,nil,nil,nil,"product")
	//test
	code, err := ioutil.ReadFile("./testdata2/contract.wasm")
	if err != nil {
		fmt.Println("error in read file", err.Error())
		return
	}

	par := make([]Param,2)
	par[0] = Param{Ptype:"int",Pval:"20"}
	par[1] = Param{Ptype:"int",Pval:"30"}

	p := Args{Params:par}
	bytes,err:=json.Marshal(p)
	if err != nil{
		fmt.Println(err)
		t.Fatal(err.Error())
	}
	fmt.Println(string(bytes))

	input := make([]interface{}, 3)
	input[0] = "invoke"
	input[1] = "add"
	input[2] = string(bytes)

	fmt.Printf("input is %v\n", input)

	res, err := engine.CallInf(common.Address{}, code, input,nil)
	if err != nil {
		fmt.Println("call error!", err.Error())
	}
	fmt.Printf("res:%v\n", res)

	retbytes,err := engine.vm.GetPointerMemory(uint64(binary.LittleEndian.Uint32(res)))
	if err != nil{
		fmt.Println(err)
		t.Fatal("errors:" + err.Error())
	}

	fmt.Println("retbytes is " +string(retbytes))

	result := &Result{}
	json.Unmarshal(retbytes,result)

	fmt.Println(engine.vm.memory.Memory[:20])
	fmt.Println(engine.vm.memory.Memory[16384:])

	fmt.Println(string(engine.vm.memory.Memory[7:50]))

	if result.Pval != "50"{
		t.Fatal("result should be 50")
	}
}



func TestContract2(t *testing.T){
	engine := NewExecutionEngine(nil,nil,nil,nil,"product")
	//test
	code, err := ioutil.ReadFile("./testdata2/contract.wasm")
	if err != nil {
		fmt.Println("error in read file", err.Error())
		return
	}

	par := make([]Param,2)
	par[0] = Param{Ptype:"int",Pval:"20"}
	par[1] = Param{Ptype:"int",Pval:"30"}

	p := Args{Params:par}
	jbytes,err:=json.Marshal(p)
	if err != nil{
		fmt.Println(err)
		t.Fatal(err.Error())
	}
	fmt.Println(string(jbytes))

	bf := bytes.NewBufferString("add")
	bf.WriteString("|")
	bf.Write(jbytes)

	fmt.Printf("input is %v\n", bf.Bytes())

	res, err := engine.Call(common.Address{}, code, bf.Bytes())
	if err != nil {
		fmt.Println("call error!", err.Error())
	}
	fmt.Printf("res:%v\n", res)

	retbytes,err := engine.vm.GetPointerMemory(uint64(binary.LittleEndian.Uint32(res)))
	if err != nil{
		fmt.Println(err)
		t.Fatal("errors:" + err.Error())
	}

	fmt.Println("retbytes is " +string(retbytes))

	result := &Result{}
	json.Unmarshal(retbytes,result)

	fmt.Println(engine.vm.memory.Memory[:20])
	fmt.Println(engine.vm.memory.Memory[16384:])

	fmt.Println(string(engine.vm.memory.Memory[7:50]))

	if result.Pval != "50"{
		t.Fatal("result should be 50")
	}
}



func TestContract3(t *testing.T){
	engine := NewExecutionEngine(nil,nil,nil,nil,"product")
	//test
	code, err := ioutil.ReadFile("./testdata2/contract.wasm")
	if err != nil {
		fmt.Println("error in read file", err.Error())
		return
	}

	par := make([]Param,2)
	par[0] = Param{Ptype:"string",Pval:"hello "}
	par[1] = Param{Ptype:"string",Pval:"world!"}

	p := Args{Params:par}
	jbytes,err:=json.Marshal(p)
	if err != nil{
		fmt.Println(err)
		t.Fatal(err.Error())
	}
	fmt.Println(string(jbytes))

	bf := bytes.NewBufferString("concat")
	bf.WriteString("|")
	bf.Write(jbytes)

	fmt.Printf("input is %v\n", bf.Bytes())

	res, err := engine.Call(common.Address{}, code, bf.Bytes())
	if err != nil {
		fmt.Println("call error!", err.Error())
	}
	fmt.Printf("res:%v\n", res)

	retbytes,err := engine.vm.GetPointerMemory(uint64(binary.LittleEndian.Uint32(res)))
	if err != nil{
		fmt.Println(err)
		t.Fatal("errors:" + err.Error())
	}

	fmt.Println("retbytes is " +string(retbytes))

	result := &Result{}
	json.Unmarshal(retbytes,result)

	if result.Pval != "hello world!"{
		t.Fatal("the res should be 'hello world!'")
	}

}

func TestContract4(t *testing.T){
	engine := NewExecutionEngine(nil,nil,nil,nil,"product")
	//test
	code, err := ioutil.ReadFile("./testdata2/contract.wasm")
	if err != nil {
		fmt.Println("error in read file", err.Error())
		return
	}

	par := make([]Param,2)
	par[0] = Param{Ptype:"int_array",Pval:"1,2,3,4,5,6"}
	par[1] = Param{Ptype:"int_array",Pval:"7,8,9,10"}

	p := Args{Params:par}
	jbytes,err:=json.Marshal(p)
	if err != nil{
		fmt.Println(err)
		t.Fatal(err.Error())
	}
	fmt.Println(string(jbytes))

	bf := bytes.NewBufferString("sumArray")
	bf.WriteString("|")
	bf.Write(jbytes)

	fmt.Printf("input is %v\n", bf.Bytes())

	res, err := engine.Call(common.Address{}, code, bf.Bytes())
	if err != nil {
		fmt.Println("call error!", err.Error())
	}
	fmt.Printf("res:%v\n", res)

	retbytes,err := engine.vm.GetPointerMemory(uint64(binary.LittleEndian.Uint32(res)))
	if err != nil{
		fmt.Println(err)
		t.Fatal("errors:" + err.Error())
	}

	fmt.Println("retbytes is " +string(retbytes))

	result := &Result{}
	json.Unmarshal(retbytes,result)

	if result.Pval != "55"{
		t.Fatal("the res should be '55'")
	}

}

func TestRawContract(t *testing.T){
	engine := NewExecutionEngine(nil,nil,nil,nil,"product")
	//test
	code, err := ioutil.ReadFile("./testdata2/rawcontract.wasm")
	if err != nil {
		fmt.Println("error in read file", err.Error())
		return
	}
	bf := bytes.NewBufferString("add")
	bf.WriteString("|")

	tmp:=make([]byte,8)
	binary.LittleEndian.PutUint32(tmp[:4],uint32(10))
	binary.LittleEndian.PutUint32(tmp[4:],uint32(20))
	bf.Write(tmp)

	fmt.Printf("input is %v\n", bf.Bytes())

	res, err := engine.Call(common.Address{}, code, bf.Bytes())
	if err != nil {
		fmt.Println("call error!", err.Error())
	}
	fmt.Printf("res:%v\n", res)

	retbytes,err := engine.vm.GetPointerMemory(uint64(binary.LittleEndian.Uint32(res)))
	if err != nil{
		fmt.Println(err)
		t.Fatal("errors:" + err.Error())
	}

	fmt.Println("retbytes is " +string(retbytes))

	result := &Result{}
	json.Unmarshal(retbytes,result)

	if result.Pval != "30"{
		t.Fatal("the res should be '30'")
	}

}

func TestRawContract2(t *testing.T){
	engine := NewExecutionEngine(nil,nil,nil,nil,"product")
	//test
	code, err := ioutil.ReadFile("./testdata2/rawcontract.wasm")
	if err != nil {
		fmt.Println("error in read file", err.Error())
		return
	}
	bf := bytes.NewBufferString("concat")
	bf.WriteString("|")

	tmp:=bytes.NewBuffer(nil)
	serialization.WriteVarString(tmp,"hello ")
	bf.Write(tmp.Bytes())

	tmp=bytes.NewBuffer(nil)
	serialization.WriteVarString(tmp,"world!")
	bf.Write(tmp.Bytes())

	fmt.Printf("input is %v\n", bf.Bytes())

	res, err := engine.Call(common.Address{}, code, bf.Bytes())
	if err != nil {
		fmt.Println("call error!", err.Error())
		t.Fatal("errors:" + err.Error())
	}
	fmt.Printf("res:%v\n", res)

	retbytes,err := engine.vm.GetPointerMemory(uint64(binary.LittleEndian.Uint32(res)))
	if err != nil{
		fmt.Println(err)
		t.Fatal("errors:" + err.Error())
	}

	fmt.Println("retbytes is " +string(retbytes))

	result := &Result{}
	json.Unmarshal(retbytes,result)

	if result.Pval != "hello world!"{
		t.Fatal("the res should be 'hello world!'")
	}

}


func TestRawContract3(t *testing.T){
	engine := NewExecutionEngine(nil,nil,nil,nil,"product")
	//test
	code, err := ioutil.ReadFile("./testdata2/rawcontract2.wasm")
	if err != nil {
		fmt.Println("error in read file", err.Error())
		return
	}
	bf := bytes.NewBufferString("concat")
	bf.WriteString("|")

	tmp:=bytes.NewBuffer(nil)
	serialization.WriteVarString(tmp,"hello ")
	bf.Write(tmp.Bytes())

	tmp=bytes.NewBuffer(nil)
	serialization.WriteVarString(tmp,"world!")
	bf.Write(tmp.Bytes())

	fmt.Printf("input is %v\n", bf.Bytes())

	res, err := engine.Call(common.Address{}, code, bf.Bytes())
	if err != nil {
		fmt.Println("call error!", err.Error())
		t.Fatal("errors:" + err.Error())
	}
	fmt.Printf("res:%v\n", res)

	retbytes,err := engine.vm.GetPointerMemory(uint64(binary.LittleEndian.Uint32(res)))
	if err != nil{
		fmt.Println(err)
		t.Fatal("errors:" + err.Error())
	}

	fmt.Println("retbytes is " +string(retbytes))

	result := &Result{}
	json.Unmarshal(retbytes,result)

	if result.Pval != "hello world!"{
		t.Fatal("the res should be 'hello world!'")
	}

}



func TestRawContract4(t *testing.T){
	engine := NewExecutionEngine(nil,nil,nil,nil,"product")
	//test
	code, err := ioutil.ReadFile("./testdata2/rawcontract2.wasm")
	if err != nil {
		fmt.Println("error in read file", err.Error())
		return
	}
	bf := bytes.NewBufferString("add")
	bf.WriteString("|")

	tmp:=make([]byte,8)
	binary.LittleEndian.PutUint32(tmp[:4],uint32(10))
	binary.LittleEndian.PutUint32(tmp[4:],uint32(20))
	bf.Write(tmp)

	fmt.Printf("input is %v\n", bf.Bytes())

	res, err := engine.Call(common.Address{}, code, bf.Bytes())
	if err != nil {
		fmt.Println("call error!", err.Error())
	}
	fmt.Printf("res:%v\n", res)

	retbytes,err := engine.vm.GetPointerMemory(uint64(binary.LittleEndian.Uint32(res)))
	if err != nil{
		fmt.Println(err)
		t.Fatal("errors:" + err.Error())
	}

	fmt.Println("retbytes is " +string(retbytes))

	result := &Result{}
	json.Unmarshal(retbytes,result)

	if result.Pval != "30"{
		t.Fatal("the res should be '30'")
	}

}
