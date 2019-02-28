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

package nodeinfo

import "strings"

type NgbNodeInfo struct {
	NgbId         string //neighbor node id
	NgbType       string
	NgbAddr       string
	HttpInfoAddr  string
	HttpInfoPort  uint16
	HttpInfoStart bool
	NgbVersion    string
}

type NgbNodeInfoSlice []NgbNodeInfo

func (n NgbNodeInfoSlice) Len() int {
	return len(n)
}

func (n NgbNodeInfoSlice) Swap(i, j int) {
	n[i], n[j] = n[j], n[i]
}

func (n NgbNodeInfoSlice) Less(i, j int) bool {
	if 0 <= strings.Compare(n[i].HttpInfoAddr, n[j].HttpInfoAddr) {
		return false
	} else {
		return true
	}
}
