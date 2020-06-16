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

import (
	"math/rand"
	"testing"

	"github.com/ontio/ontology/common/config"
	"github.com/ontio/ontology/common/constants"
	"github.com/stretchr/testify/assert"
)

func TestCalcUnbindOng(t *testing.T) {
	assert.Equal(t, CalcUnbindOng(1, 0, 1), uint64(GENERATION_AMOUNT[0]))
	assert.Equal(t, CalcUnbindOng(1, 0, TIME_INTERVAL), GENERATION_AMOUNT[0]*uint64(TIME_INTERVAL))
	assert.Equal(t, CalcUnbindOng(1, 0, TIME_INTERVAL+1),
		GENERATION_AMOUNT[1]+GENERATION_AMOUNT[0]*uint64(TIME_INTERVAL))
	assert.Equal(t, CalcUnbindOng(1, config.GetOntHolderUnboundDeadline(),
		config.GetOntHolderUnboundDeadline()+1), uint64(0))
	assert.Equal(t, CalcUnbindOng(1, config.GetOntHolderUnboundDeadline()-1,
		config.GetOntHolderUnboundDeadline()+1), uint64(3))
	assert.Equal(t, CalcUnbindOng(1, config.GetOntHolderUnboundDeadline()-2,
		config.GetOntHolderUnboundDeadline()+1), uint64(2*3))
	assert.Equal(t, CalcUnbindOng(1, config.GetOntHolderUnboundDeadline()-2,
		config.GetOntHolderUnboundDeadline()), uint64(2*3))
}

// test identity: unbound[t1, t3) = unbound[t1, t2) + unbound[t2, t3)
func TestCumulative(t *testing.T) {
	N := 10000
	for i := 0; i < N; i++ {
		tstart := rand.Uint32()
		tend := tstart + rand.Uint32()
		tmid := uint32((uint64(tstart) + uint64(tend)) / 2)

		total := CalcUnbindOng(1, tstart, tend)
		total2 := CalcUnbindOng(1, tstart, tmid) + CalcUnbindOng(1, tmid, tend)
		assert.Equal(t, total, total2)
	}
}

// test 1 balance will not get ONT_TOTAL_SUPPLY eventually
func TestTotalONG(t *testing.T) {
	total := CalcUnbindOng(constants.ONT_TOTAL_SUPPLY, 0, TIME_INTERVAL*18) + CalcGovernanceUnbindOng(0, TIME_INTERVAL*18)
	assert.Equal(t, total, constants.ONG_TOTAL_SUPPLY)

	total = CalcUnbindOng(constants.ONT_TOTAL_SUPPLY, 0, TIME_INTERVAL*108) + CalcGovernanceUnbindOng(0, TIME_INTERVAL*108)
	assert.Equal(t, total, constants.ONG_TOTAL_SUPPLY)

	total = CalcUnbindOng(constants.ONT_TOTAL_SUPPLY, 0, ^uint32(0)) + CalcGovernanceUnbindOng(0, ^uint32(0))
	assert.Equal(t, total, constants.ONG_TOTAL_SUPPLY)
}

func TestCalcGovernanceUnbindOng(t *testing.T) {
	assert.Equal(t, CalcGovernanceUnbindOng(0, 1), uint64(0))
	assert.Equal(t, CalcGovernanceUnbindOng(0, TIME_INTERVAL), uint64(0))
	assert.Equal(t, CalcGovernanceUnbindOng(0, TIME_INTERVAL+1), uint64(0))
	assert.Equal(t, CalcGovernanceUnbindOng(config.GetOntHolderUnboundDeadline(),
		config.GetOntHolderUnboundDeadline()+1), uint64(constants.ONT_TOTAL_SUPPLY))
	assert.Equal(t, CalcGovernanceUnbindOng(config.GetOntHolderUnboundDeadline()-1,
		config.GetOntHolderUnboundDeadline()+1), uint64(constants.ONT_TOTAL_SUPPLY))
	assert.Equal(t, CalcGovernanceUnbindOng(config.GetOntHolderUnboundDeadline()-2,
		config.GetOntHolderUnboundDeadline()+1), uint64(constants.ONT_TOTAL_SUPPLY))
	assert.Equal(t, CalcGovernanceUnbindOng(config.GetOntHolderUnboundDeadline()-2,
		config.GetOntHolderUnboundDeadline()), uint64(0))
	assert.Equal(t, CalcGovernanceUnbindOng(config.GetOntHolderUnboundDeadline()-2,
		config.GetOntHolderUnboundDeadline()+2), uint64(2*constants.ONT_TOTAL_SUPPLY))
}
