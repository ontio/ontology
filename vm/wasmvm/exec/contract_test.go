package exec

import (
	"testing"
	"io/ioutil"
	"fmt"
	"github.com/Ontology/common"
	"encoding/json"
	"encoding/binary"
)




func TestContract1(t *testing.T){
	engine := NewExecutionEngine(nil,nil,nil,nil)
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

	res, err := engine.CallInf(common.Uint160{}, code, input,nil)
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

	if result.Pval != "50"{
		t.Fatal("result should be 50")
	}
}
