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

package states

import (
	"bytes"
	"io"

	"github.com/ontio/ontology/common"
	"github.com/ontio/ontology/common/serialization"
)

type StorageKey struct {
	ContractAddress common.Address
	Key             []byte
}

func (this *StorageKey) Serialize(w io.Writer) (int, error) {
	if err := this.ContractAddress.Serialize(w); err != nil {
		return 0, err
	}
	if err := serialization.WriteVarBytes(w, this.Key); err != nil {
		return 0, err
	}
	return 0, nil
}

func (this *StorageKey) Deserialize(r io.Reader) error {
	if err := this.ContractAddress.Deserialize(r); err != nil {
		return err
	}
	key, err := serialization.ReadVarBytes(r)
	if err != nil {
		return err
	}
	this.Key = key
	return nil
}

func (this *StorageKey) ToArray() []byte {
	b := new(bytes.Buffer)
	this.Serialize(b)
	return b.Bytes()
}
