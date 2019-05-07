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
package testsuite

import (
	"github.com/ontio/ontology/core/store/common"
	"github.com/ontio/ontology/core/store/overlaydb"
)

type FakeDB struct {
	common.PersistStore
}

func (self *FakeDB) Get(key []byte) ([]byte, error) {
	return nil, common.ErrNotFound
}

func NewOverlayDB() *overlaydb.OverlayDB {
	return overlaydb.NewOverlayDB(&FakeDB{nil})
}
