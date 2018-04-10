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

package vbft

import (
	"testing"
)

func TestSerialize(t *testing.T) {
	blk, err := constructBlock()
	if err != nil {
		t.Errorf("constructBlock failed: %v", err)
	}
	_, err = blk.Serialize()
	if err != nil {
		t.Errorf("Block Serialize failed :%v", err)
		return
	}
	t.Log("Block Serialize succ")
}

func TestInitVbftBlock(t *testing.T) {
	blk, err := constructBlock()
	if err != nil {
		t.Errorf("constructBlock failed: %v", err)
	}
	_, err = initVbftBlock(blk.Block)
	if err != nil {
		t.Errorf("initVbftBlock failed: %v", err)
		return
	}
	t.Log("TestInitVbftBlock succ")
}
