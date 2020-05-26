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

package subnet

import "net"

type SubNetReservedAddrFilter struct {
	subnet *SubNet
}

func (self *SubNetReservedAddrFilter) Filtered(addr string) bool {
	// seed node should allow all node connection
	if self.subnet.seedNode {
		return false
	}

	ip, _, err := net.SplitHostPort(addr)
	if err != nil {
		return true
	}

	if self.subnet.isSeedIp(ip) || self.subnet.acct == nil || !self.subnet.gov.IsGovNode(self.subnet.acct.PublicKey) {
		return false
	}

	// self is gov node, then check whether addr is subnet members
	has := self.subnet.IpInMembers(ip)
	return !has
}

type SubNetMaskAddrFilter struct {
	subnet *SubNet
}

func (self *SubNetMaskAddrFilter) Filtered(addr string) bool {
	self.subnet.lock.Lock()
	defer self.subnet.lock.Unlock()
	_, ok := self.subnet.members[addr]

	return ok
}
