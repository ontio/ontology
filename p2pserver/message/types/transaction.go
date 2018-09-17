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

package types

import (
	comm "github.com/ontio/ontology/common"
	"github.com/ontio/ontology/core/types"
	"github.com/ontio/ontology/p2pserver/common"
)

// Transaction message
type Trn struct {
	Txn *types.Transaction
}

//Serialize message payload
func (this Trn) Serialization(sink *comm.ZeroCopySink) error {
	return this.Txn.Serialization(sink)
}

func (this *Trn) CmdType() string {
	return common.TX_TYPE
}

//Deserialize message payload
func (this *Trn) Deserialization(source *comm.ZeroCopySource) error {
	tx := &types.Transaction{}
	err := tx.Deserialization(source)
	if err != nil {
		return err
	}

	this.Txn = tx
	return nil
}
