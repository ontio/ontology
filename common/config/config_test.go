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
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestConfigGeneration(t *testing.T) {
	assert.NotNil(t, DefConfig)
	assert.NotNil(t, DefConfig.Genesis)
	assert.NotNil(t, DefConfig.Common)
	assert.NotNil(t, DefConfig.Consensus)
	assert.NotNil(t, DefConfig.P2PNode)
	assert.NotNil(t, DefConfig.Rpc)
	assert.NotNil(t, DefConfig.Ws)
	assert.NotNil(t, DefConfig.Cli)

	assert.Equal(t, DefConfig.Genesis, PolarisConfig)

	assert.Equal(t, DefConfig.Common.GasLimit, uint64(DEFAULT_GAS_LIMIT))
	assert.Equal(t, DefConfig.Common.GasPrice, uint64(DEFAULT_GAS_PRICE))

	assert.Equal(t, DefConfig.Consensus.EnableConsensus, true)
	assert.Equal(t, DefConfig.Consensus.MaxTxInBlock, uint(DEFAULT_MAX_TX_IN_BLOCK))

	assert.Equal(t, DefConfig.P2PNode.NodePort, DEFAULT_NODE_PORT)
	assert.Equal(t, DefConfig.P2PNode.NetworkId, uint(DEFAULT_NET_MAGIC))

	assert.Equal(t, DefConfig.Rpc.HttpJsonPort, DEFAULT_RPC_PORT)
	assert.Equal(t, DefConfig.Rpc.HttpLocalPort, DEFAULT_RPC_LOCAL_PORT)
}
