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
	"bytes"
	"fmt"

	"github.com/ontio/ontology/common/serialization"
	"github.com/ontio/ontology/errors"
	"github.com/ontio/ontology/p2pserver/common"
)

type VerACK struct {
	IsConsensus bool
}

//Serialize message payload
func (this VerACK) Serialization() ([]byte, error) {
	p := bytes.NewBuffer([]byte{})
	err := serialization.WriteBool(p, this.IsConsensus)
	if err != nil {
		return nil, errors.NewDetailErr(err, errors.ErrNetPackFail, fmt.Sprintf("write error. IsConsensus:%v", this.IsConsensus))
	}

	return p.Bytes(), nil
}

func (this VerACK) CmdType() string {
	return common.VERACK_TYPE
}

//Deserialize message payload
func (this *VerACK) Deserialization(p []byte) error {
	buf := bytes.NewBuffer(p)

	isConsensus, err := serialization.ReadBool(buf)
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNetUnPackFail, fmt.Sprintf("read IsConsensus error. buf:%v", buf))
	}

	this.IsConsensus = isConsensus
	return nil
}
