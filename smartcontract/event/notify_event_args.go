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

package event

import (
	"github.com/ontio/ontology/common"
	"github.com/ontio/ontology/vm/neovm/types"
)

// NotifyEventArgs describe smart contract event notify arguments struct
type NotifyEventArgs struct {
	TxHash          common.Uint256
	ContractAddress common.Address
	States          types.StackItems
}

// NotifyEventInfo describe smart contract event notify info struct
type NotifyEventInfo struct {
	TxHash          common.Uint256
	ContractAddress common.Address
	States          interface{}
}
