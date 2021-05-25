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

package ethl2

import (
	"github.com/ontio/ontology/common"
	"github.com/ontio/ontology/common/config"
	"github.com/ontio/ontology/smartcontract/event"
	"github.com/ontio/ontology/smartcontract/service/native"
	"github.com/ontio/ontology/smartcontract/service/native/utils"
)

type State struct {
	fName string
	ethtx []byte
}

func (s *State) Serialization(sink *common.ZeroCopySink) {
	utils.EncodeVarBytes(sink, s.ethtx)
}

func (s *State) Deserialization(source *common.ZeroCopySource) error {
	var err error

	s.ethtx, err = utils.DecodeVarBytes(source)

	return err
}

func AddNotifications(native *native.NativeService, contract common.Address, state *State) {
	if !config.DefConfig.Common.EnableEventLog {
		return
	}

	native.Notifications = append(native.Notifications,
		&event.NotifyEventInfo{
			ContractAddress: contract,
			States:          []interface{}{state.fName, state.ethtx},
		})
}

func AddAppendAddressNotification(native *native.NativeService, contract common.Address, addrs []common.Address) {
	if !config.DefConfig.Common.EnableEventLog {
		return
	}
	lst := []interface{}{MethodAppendAddress}

	for _, addr := range addrs {
		lst = append(lst, addr.ToBase58())
	}
	noti := &event.NotifyEventInfo{
		ContractAddress: contract,
		States:          lst,
	}

	native.Notifications = append(native.Notifications, noti)
}

func AddNotification(native *native.NativeService, contract common.Address, args ...interface{}) {
	if !config.DefConfig.Common.EnableEventLog {
		return
	}

	noti := &event.NotifyEventInfo{
		ContractAddress: contract,
		States:          args,
	}

	native.Notifications = append(native.Notifications, noti)
}
