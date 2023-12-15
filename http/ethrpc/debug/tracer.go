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

package debug

import (
	"fmt"

	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/rpc"
	"github.com/ontio/ontology/core/ledger"
	"github.com/ontio/ontology/http/ethrpc/eth"
	types2 "github.com/ontio/ontology/http/ethrpc/types"
	"github.com/ontio/ontology/vm/evm"
	"github.com/ontio/ontology/vm/evm/tracers"
)

// DebugAPI is the collection of tracing APIs exposed over the private debugging endpoint.
type DebugAPI struct {
}

// NewDebugAPI creates a new DebugAPI definition for the tracing methods of the Ethereum service.
func NewDebugAPI() *DebugAPI {
	return &DebugAPI{}
}

// TraceConfig holds extra parameters to trace functions.
type TraceConfig struct {
	*evm.LogConfig
	Tracer  *string
	Timeout *string
	Reexec  *uint64
}

// TraceCallConfig is the config for traceCall DebugAPI. It holds one more
// field to override the state for tracing.
type TraceCallConfig struct {
	*evm.LogConfig
	Tracer  *string
	Timeout *string
	Reexec  *uint64
	//StateOverrides *ethapi.StateOverride
}

// TraceCall lets you trace a given eth_call. It collects the structured logs
// created during the execution of EVM if the given transaction was added on
// top of the provided block and returns them as a JSON object.
// You can provide -2 as a block number to trace on top of the pending block.
func (api *DebugAPI) TraceCall(args types2.CallArgs, blockNrOrHash rpc.BlockNumberOrHash, config *TraceCallConfig) (interface{}, error) {
	// Execute the trace
	msg := args.AsMessage(eth.RPCGasCap)

	var traceConfig *TraceConfig
	if config != nil {
		traceConfig = &TraceConfig{
			LogConfig: config.LogConfig,
			Tracer:    config.Tracer,
			Timeout:   config.Timeout,
			Reexec:    config.Reexec,
		}
	}
	return api.traceTx(msg, traceConfig)
}

// traceTx configures a new tracer according to the provided configuration, and
// executes the given message in the provided environment. The return value will
// be tracer dependent.
func (api *DebugAPI) traceTx(message types.Message, config *TraceConfig) (interface{}, error) {
	// Assemble the structured logger or the JavaScript tracer
	var (
		tracer evm.Tracer
		err    error
	)
	switch {
	case config == nil:
		tracer = evm.NewStructLogger(nil)
	case config.Tracer != nil:
		switch *config.Tracer {
		case "callTracer":
			tracer = tracers.NewCallTracer()
		default:
			return nil, fmt.Errorf("unkown tracer type: %s", *config.Tracer)
		}
	default:
		tracer = evm.NewStructLogger(config.LogConfig)
	}

	result, err := ledger.DefLedger.TraceEip155Tx(message, nil)
	if err != nil {
		return nil, fmt.Errorf("tracing failed: %w", err)
	}

	// Depending on the tracer type, format and return the output.
	switch tracer := tracer.(type) {
	case *evm.StructLogger:
		// If the result contains a revert reason, return it.
		returnVal := fmt.Sprintf("%x", result.Return())
		if len(result.Revert()) > 0 {
			returnVal = fmt.Sprintf("%x", result.Revert())
		}
		return &ExecutionResult{
			Gas:         result.UsedGas,
			Failed:      result.Failed(),
			ReturnValue: returnVal,
			StructLogs:  FormatLogs(tracer.StructLogs()),
		}, nil

	case *tracers.CallTracer:
		return tracer.GetResult()

	default:
		panic(fmt.Sprintf("bad tracer type %T", tracer))
	}
}
