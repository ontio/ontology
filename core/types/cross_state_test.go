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
package types

import (
	"testing"

	"github.com/ontio/ontology/common"
	"github.com/stretchr/testify/assert"
)

func TestCrossState(t *testing.T) {
	sigData := make([][]byte, 3)
	sigData[0] = []byte{1, 2, 3, 4, 5}
	sigData[1] = []byte{2, 3, 4, 5, 6}
	sigData[2] = []byte{3, 4, 5, 6, 7}

	msg := &CrossChainMsg{
		Version:    CURR_CROSS_STATES_VERSION,
		Height:     1,
		StatesRoot: common.UINT256_EMPTY,
		SigData:    sigData,
	}
	sink := common.NewZeroCopySink(nil)
	msg.Serialization(sink)

	source := common.NewZeroCopySource(sink.Bytes())

	var msg1 CrossChainMsg
	err := msg1.Deserialization(source)

	assert.NoError(t, err)
	assert.Equal(t, *msg, msg1)
}
