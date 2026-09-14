// Copyright (C) 2024 The ontology Authors
// This file is part of the ontology library.
//
// The ontology is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The ontology is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.

package js

import (
	"math/big"
	"strings"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ontio/ontology/http/ethrpc/tracers"
	"github.com/ontio/ontology/vm/evm"
	"github.com/ontio/ontology/vm/evm/params"
)

// dummyStateDB embeds the StateDB interface so that only the methods exercised
// while driving the capture hooks need to be implemented.
type dummyStateDB struct {
	evm.StateDB
}

func (d *dummyStateDB) GetRefund() uint64 { return 0 }

func testEVM() *evm.EVM {
	blockCtx := evm.BlockContext{
		BlockNumber: big.NewInt(1),
		GetHash:     func(uint64) common.Hash { return common.Hash{} },
	}
	txCtx := evm.TxContext{GasPrice: big.NewInt(1)}
	return evm.NewEVM(blockCtx, txCtx, &dummyStateDB{}, params.TestChainConfig, evm.Config{})
}

// TestJsTracerInitTimeout verifies that a tracer whose top level code never
// returns is interrupted instead of hanging the request handler.
func TestJsTracerInitTimeout(t *testing.T) {
	start := time.Now()
	_, err := newJsTracerWithTimeout("(function(){ while(true){} })()", new(tracers.Context), nil, 200*time.Millisecond)
	if err == nil {
		t.Fatal("expected tracer initialization to time out")
	}
	if !strings.Contains(err.Error(), "execution timeout") {
		t.Fatalf("expected timeout error, got %v", err)
	}
	if elapsed := time.Since(start); elapsed > 5*time.Second {
		t.Fatalf("tracer initialization was not interrupted in time: %v", elapsed)
	}
}

// TestJsTracerResultTimeout verifies that a tracer that spins forever inside
// its result() function is interrupted.
func TestJsTracerResultTimeout(t *testing.T) {
	tracer, err := newJsTracerWithTimeout(`{fault: function(){}, result: function(){ while(true){} }}`,
		new(tracers.Context), nil, 200*time.Millisecond)
	if err != nil {
		t.Fatal(err)
	}
	start := time.Now()
	_, err = tracer.GetResult()
	if err == nil {
		t.Fatal("expected result generation to time out")
	}
	if !strings.Contains(err.Error(), "execution timeout") {
		t.Fatalf("expected timeout error, got %v", err)
	}
	if elapsed := time.Since(start); elapsed > 5*time.Second {
		t.Fatalf("result generation was not interrupted in time: %v", elapsed)
	}
}

// TestJsTracerStepTimeout verifies that a tracer that spins forever in a step()
// hook invoked during EVM execution is interrupted and that the EVM is asked to
// stop.
func TestJsTracerStepTimeout(t *testing.T) {
	tracer, err := newJsTracerWithTimeout(`{step: function(log){ while(true){} }, fault: function(){}, result: function(){ return null; }}`,
		new(tracers.Context), nil, 200*time.Millisecond)
	if err != nil {
		t.Fatal(err)
	}
	js := tracer.(*jsTracer)
	env := testEVM()
	js.CaptureStart(env, common.Address{}, common.Address{}, false, nil, 100000, big.NewInt(0))

	done := make(chan struct{})
	go func() {
		js.CaptureState(0, evm.OpCode(0), 100000, 0, &evm.ScopeContext{}, nil, 0, nil)
		close(done)
	}()
	select {
	case <-done:
	case <-time.After(5 * time.Second):
		t.Fatal("step hook was not interrupted by the tracer timeout")
	}
	if !env.Cancelled() {
		t.Fatal("expected EVM to be cancelled after tracer timeout")
	}
	if js.err == nil {
		t.Fatal("expected step error to be recorded")
	}
}

// TestJsTracerDefaultTimeout verifies that a tracer created without an explicit
// timeout falls back to the package default.
func TestJsTracerDefaultTimeout(t *testing.T) {
	tracer, err := newJsTracer(`{fault: function(){}, result: function(){ return null; }}`, new(tracers.Context), nil)
	if err != nil {
		t.Fatal(err)
	}
	js := tracer.(*jsTracer)
	if js.timeout != tracers.DefaultTraceTimeout {
		t.Fatalf("expected default timeout %v, got %v", tracers.DefaultTraceTimeout, js.timeout)
	}
	if _, err := tracer.GetResult(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}
