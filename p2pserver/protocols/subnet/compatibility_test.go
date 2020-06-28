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

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCompatibility(t *testing.T) {
	unsupported := []string{"v1.10.0", "1.10.0-alpha", "v1.9"}
	for _, version := range unsupported {
		assert.False(t, supportSubnet(version))
	}

	supported := []string{"v2.0.0-0-gfcbf82c", "v2.0.0", "2.0.0-alpha", "2.0.0-beta", "v2.0.0-alpha.9", "v2.0.0-laizy", "v2.0.0-laizy1"}
	for _, version := range supported {
		assert.True(t, supportSubnet(version))
	}
}
