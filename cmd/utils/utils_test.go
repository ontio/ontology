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
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestFormatOng(t *testing.T) {
	assert.Equal(t, "1", FormatOng(1000000000))
	assert.Equal(t, "1.1", FormatOng(1100000000))
	assert.Equal(t, "1.123456789", FormatOng(1123456789))
	assert.Equal(t, "1000000000.123456789", FormatOng(1000000000123456789))
	assert.Equal(t, "1000000000.000001", FormatOng(1000000000000001000))
	assert.Equal(t, "1000000000.000000001", FormatOng(1000000000000000001))
}

func TestParseOng(t *testing.T) {
	assert.Equal(t, uint64(1000000000), ParseOng("1"))
	assert.Equal(t, uint64(1000000000000000000), ParseOng("1000000000"))
	assert.Equal(t, uint64(1000000000123456789), ParseOng("1000000000.123456789"))
	assert.Equal(t, uint64(1000000000000000100), ParseOng("1000000000.0000001"))
	assert.Equal(t, uint64(1000000000000000001), ParseOng("1000000000.000000001"))
	assert.Equal(t, uint64(1000000000000000001), ParseOng("1000000000.000000001123"))
}

func TestFormatOnt(t *testing.T) {
	assert.Equal(t, "0", FormatOnt(0))
	assert.Equal(t, "1", FormatOnt(1))
	assert.Equal(t, "100", FormatOnt(100))
	assert.Equal(t, "1000000000", FormatOnt(1000000000))
}

func TestParseOnt(t *testing.T) {
	assert.Equal(t, uint64(0), ParseOnt("0"))
	assert.Equal(t, uint64(1), ParseOnt("1"))
	assert.Equal(t, uint64(1000), ParseOnt("1000"))
	assert.Equal(t, uint64(1000000000), ParseOnt("1000000000"))
	assert.Equal(t, uint64(1000000), ParseOnt("1000000.123"))
}

func TestGenExportBlocksFileName(t *testing.T) {
	name := "blocks.dat"
	start := uint32(0)
	end := uint32(100)
	fileName := GenExportBlocksFileName(name, start, end)
	assert.Equal(t, "blocks_0_100.dat", fileName)
	name = "blocks"
	fileName = GenExportBlocksFileName(name, start, end)
	assert.Equal(t, "blocks_0_100", fileName)
	name = "blocks."
	fileName = GenExportBlocksFileName(name, start, end)
	assert.Equal(t, "blocks_0_100.", fileName)
	name = "blocks.export.dat"
	fileName = GenExportBlocksFileName(name, start, end)
	assert.Equal(t, "blocks.export_0_100.dat", fileName)
}
