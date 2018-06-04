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

package utils

import "github.com/ontio/ontology/common/constants"

var (
	TIME_INTERVAL     = constants.UNBOUND_TIME_INTERVAL
	GENERATION_AMOUNT = constants.UNBOUND_GENERATION_AMOUNT
)

// startOffset : start timestamp offset from genesis block
// endOffset :  end timestamp offset from genesis block
func CalcUnbindOng(balance uint64, startOffset, endOffset uint32) uint64 {
	var amount uint64 = 0
	if startOffset >= endOffset {
		return 0
	}
	if startOffset < constants.UNBOUND_DEADLINE {
		ustart := startOffset / TIME_INTERVAL
		istart := startOffset % TIME_INTERVAL
		if endOffset >= constants.UNBOUND_DEADLINE {
			endOffset = constants.UNBOUND_DEADLINE
		}
		uend := endOffset / TIME_INTERVAL
		iend := endOffset % TIME_INTERVAL
		for ustart < uend {
			amount += uint64(TIME_INTERVAL-istart) * GENERATION_AMOUNT[ustart]
			ustart++
			istart = 0
		}
		amount += uint64(iend-istart) * GENERATION_AMOUNT[ustart]
	}

	return uint64(amount) * balance
}
