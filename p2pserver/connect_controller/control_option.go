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
package connect_controller

import (
	"github.com/ontio/ontology/common/config"
	p2p "github.com/ontio/ontology/p2pserver/net/protocol"
)

type ConnCtrlOption struct {
	MaxConnOutBound     uint
	MaxConnInBound      uint
	MaxConnInBoundPerIP uint
	ReservedPeers       p2p.AddressFilter // enabled if not empty
	dialer              Dialer
}

func NewConnCtrlOption() ConnCtrlOption {
	return ConnCtrlOption{
		MaxConnInBound:      config.DEFAULT_MAX_CONN_IN_BOUND,
		MaxConnOutBound:     config.DEFAULT_MAX_CONN_OUT_BOUND,
		MaxConnInBoundPerIP: config.DEFAULT_MAX_CONN_IN_BOUND_FOR_SINGLE_IP,
		ReservedPeers:       p2p.AllAddrFilter(),
		dialer:              &noTlsDialer{},
	}
}

func (self ConnCtrlOption) MaxOutBound(n uint) ConnCtrlOption {
	self.MaxConnOutBound = n
	return self
}

func (self ConnCtrlOption) MaxInBound(n uint) ConnCtrlOption {
	self.MaxConnInBound = n
	return self
}

func (self ConnCtrlOption) MaxInBoundPerIp(n uint) ConnCtrlOption {
	self.MaxConnInBoundPerIP = n
	return self
}

func (self ConnCtrlOption) ReservedOnly(peers []string) ConnCtrlOption {
	self.ReservedPeers = NewStaticReserveFilter(peers)
	return self
}

func (self ConnCtrlOption) WithDialer(dialer Dialer) ConnCtrlOption {
	self.dialer = dialer
	return self
}

func ConnCtrlOptionFromConfig(config *config.P2PNodeConfig, reserveFilter p2p.AddressFilter) (option ConnCtrlOption, err error) {
	dialer, e := NewDialer(config)
	if e != nil {
		err = e
		return
	}
	return ConnCtrlOption{
		MaxConnOutBound:     config.MaxConnOutBound,
		MaxConnInBound:      config.MaxConnInBound,
		MaxConnInBoundPerIP: config.MaxConnInBoundForSingleIP,
		ReservedPeers:       reserveFilter,

		dialer: dialer,
	}, nil
}
