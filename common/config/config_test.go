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

package config

import (
	"testing"

	"github.com/ontio/ontology/common/constants"
	"github.com/stretchr/testify/assert"
)

func TestGetNewUnboundDeadline(t *testing.T) {
	expected := (5 + 4) * constants.UNBOUND_TIME_INTERVAL
	expected += 3 * (GetChangeUnboundTimestamp() - constants.GENESIS_BLOCK_TIMESTAMP - 2*constants.UNBOUND_TIME_INTERVAL)
	expected += 1 * (3*constants.UNBOUND_TIME_INTERVAL - (GetChangeUnboundTimestamp() - constants.GENESIS_BLOCK_TIMESTAMP))
	for i := 3; i < len(constants.NEW_UNBOUND_GENERATION_AMOUNT); i++ {
		expected += uint32(constants.NEW_UNBOUND_GENERATION_AMOUNT[i]) * constants.UNBOUND_TIME_INTERVAL
	}
	expected = constants.UNBOUND_TIME_INTERVAL*18 - (expected - uint32(constants.ONT_TOTAL_SUPPLY))
	assert.Equal(t, expected, GetNewUnboundDeadline())
}
