package main

import (
	"fmt"
	"github.com/ontio/ontology/smartcontract"
	"testing"
)

func TestStructClone(t *testing.T) {

	config := &smartcontract.Config{
		Time:   10,
		Height: 10,
		Tx:     nil,
	}
	sc := smartcontract.SmartContract{
		Config:     config,
		Gas:        100,
		CloneCache: nil,
	}
	engine, err := sc.NewExecuteEngine(byteCode)
	if err != nil {
		panic(err)
	}
	_, err = engine.Invoke()
	if err != nil {
		panic(err)
	}
}
