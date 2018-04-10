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

package payload

import (
	"fmt"
	"io"

	"github.com/ontio/ontology/common"
	stypes "github.com/ontio/ontology/smartcontract/types"
)

type InvokeCode struct {
	GasLimit common.Fixed64
	Code     stypes.VmCode
}

func (self *InvokeCode) Serialize(w io.Writer) error {
	var err error
	err = self.GasLimit.Serialize(w)
	if err != nil {
		return fmt.Errorf("InvokeCode GasLimit Serialize failed: %s", err)
	}
	err = self.Code.Serialize(w)
	if err != nil {
		return fmt.Errorf("InvokeCode Code Serialize failed: %s", err)
	}
	return err
}

func (self *InvokeCode) Deserialize(r io.Reader) error {
	var err error

	err = self.GasLimit.Deserialize(r)
	if err != nil {
		return fmt.Errorf("InvokeCode GasLimit Deserialize failed: %s", err)
	}
	err = self.Code.Deserialize(r)
	if err != nil {
		return fmt.Errorf("InvokeCode Code Deserialize failed: %s", err)
	}
	return nil
}
