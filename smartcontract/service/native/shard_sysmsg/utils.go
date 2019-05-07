/*
 * Copyright (C) 2019 The ontology Authors
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

package shardsysmsg

import (
	"bytes"
	"fmt"
	"github.com/ontio/ontology/smartcontract/service/native/utils"
	"sort"

	"github.com/ontio/ontology/common"
	"github.com/ontio/ontology/core/chainmgr/xshard_state"
	"github.com/ontio/ontology/smartcontract/service/native"
)

func lockTxContracts(ctx *native.NativeService, tx common.Uint256, result []byte, resultErr error) error {
	if result != nil {
		// save result/err to txstate-db
		if err := xshard_state.SetTxResult(tx, result, resultErr); err != nil {
			return fmt.Errorf("save Tx result: %s", err)
		}
	}

	contracts, err := xshard_state.GetTxContracts(tx)
	if err != nil {
		return fmt.Errorf("failed to get contract of tx %v", tx)
	}
	if len(contracts) > 1 {
		sort.Slice(contracts, func(i, j int) bool {
			return bytes.Compare(contracts[i][:], contracts[j][:]) > 0
		})
	}
	for _, c := range contracts {
		if err := xshard_state.LockContract(c); err != nil {
			// TODO: revert all locks
			return fmt.Errorf("failed to lock contract %v for tx %v", c, tx)
		}
	}

	return nil
}

func unlockTxContract(ctx *native.NativeService, tx common.Uint256) error {
	contracts, err := xshard_state.GetTxContracts(tx)
	if err != nil {
		return err
	}

	for _, c := range contracts {
		xshard_state.UnlockContract(c)
	}
	return nil
}

func NotifyShard(native *native.NativeService, toShard common.ShardID, contract common.Address, method string, args []byte) error {
	paramBytes := new(bytes.Buffer)
	params := &NotifyReqParam{
		ToShard:    toShard,
		ToContract: contract,
		Method:     method,
		Args:       args,
	}
	if err := params.Serialize(paramBytes); err != nil {
		return fmt.Errorf("NotifyShard: serialize param failed, err: %s", err)
	}
	if _, err := native.NativeCall(utils.ShardSysMsgContractAddress, REMOTE_NOTIFY, paramBytes.Bytes()); err != nil {
		return fmt.Errorf("NotifyShard: invoke failed, err: %s", err)
	}
	return nil
}
