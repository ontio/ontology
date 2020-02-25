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
package states

import (
	"testing"

	"fmt"
	"github.com/magiconair/properties/assert"
	"github.com/ontio/ontology/common"
	"github.com/ontio/ontology/smartcontract/event"
)

func TestContract_Serialize_Deserialize(t *testing.T) {
	addr := common.AddressFromVmCode([]byte{1})

	c := &ContractInvokeParam{
		Version: 0,
		Address: addr,
		Method:  "init",
		Args:    []byte{2},
	}
	sink := common.NewZeroCopySink(nil)
	c.Serialization(sink)

	v := new(ContractInvokeParam)
	source := common.NewZeroCopySource(sink.Bytes())
	if err := v.Deserialization(source); err != nil {
		t.Fatalf("ContractInvokeParam deserialize error: %v", err)
	}
}

func TestPreExecResult_FromJson(t *testing.T) {
	evts := make([]*event.NotifyEventInfo, 0)
	addr, _ := common.AddressFromHexString("10679eac09c4b619685ea29c6cdc301eac8e46ed")
	evts = append(evts, &event.NotifyEventInfo{addr, "states"})
	pr := PreExecResult{
		State:  1,
		Gas:    20000,
		Result: "test",
		Notify: evts,
	}
	per, _ := pr.ToJson()
	fmt.Println(per)

	pr2 := PreExecResult{}
	pr2.FromJson([]byte(per))
	fmt.Println(pr2)
	per2, _ := pr2.ToJson()
	assert.Equal(t, per, per2)
}
