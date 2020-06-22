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

package p2p

type AddressFilter interface {
	// addr format : ip:port
	Contains(addr string) bool
}

func CombineAddrFilter(filter1, filter2 AddressFilter) AddressFilter {
	return &combineAddrFilter{filter1: filter1, filter2: filter2}
}

func NoneAddrFilter() AddressFilter {
	return &noneAddrFilter{}
}

type combineAddrFilter struct {
	filter1 AddressFilter
	filter2 AddressFilter
}

func (self *combineAddrFilter) Contains(addr string) bool {
	return self.filter1.Contains(addr) || self.filter2.Contains(addr)
}

type noneAddrFilter struct{}

func (self *noneAddrFilter) Contains(addr string) bool {
	return false
}

func AllAddrFilter() AddressFilter {
	return &allAddrFilter{}
}

type allAddrFilter struct{}

func (self *allAddrFilter) Contains(addr string) bool {
	return true
}
