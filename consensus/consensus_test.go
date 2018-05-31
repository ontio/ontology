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
package consensus

import (
	"github.com/ontio/ontology/common/log"
	"github.com/ontio/ontology/consensus/dbft"
	"github.com/ontio/ontology/consensus/solo"
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
)

func TestNewConsensusService(t *testing.T) {
	defer func() {
		os.RemoveAll(log.PATH)
		os.RemoveAll("./Chain")
	}()
	log.InitLog(log.InfoLog, log.PATH, log.Stdout)

	consensus, err := NewConsensusService("", nil, nil, nil, nil)
	assert.Nil(t, err)
	assert.NotNil(t, consensus)
	if _, ok := consensus.(*dbft.DbftService); !ok {
		t.Fatal("test new default consensus failed!")
	}

	consensus, err = NewConsensusService(CONSENSUS_DBFT, nil, nil, nil, nil)
	assert.NotNil(t, err) // spawn: name exists; because generate dbft serice twice
	assert.NotNil(t, consensus)
	if _, ok := consensus.(*dbft.DbftService); !ok {
		t.Fatal("test new dbft consensus failed!")
	}

	consensus, err = NewConsensusService(CONSENSUS_SOLO, nil, nil, nil, nil)
	assert.Nil(t, err)
	assert.NotNil(t, consensus)
	if _, ok := consensus.(*solo.SoloService); !ok {
		t.Fatal("test new solo consensus failed!")
	}

	assert.Panics(t, func() {
		NewConsensusService(CONSENSUS_VBFT, nil, nil, nil, nil)
	})
}
