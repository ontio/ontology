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

	"github.com/ontio/ontology/common/log"
	"github.com/ontio/ontology/errors"
	"github.com/ontio/ontology/p2pserver/common"
)

type Consensus struct {
	Cons ConsensusPayload
}

//Serialize message payload
func (this *Consensus) Serialization() ([]byte, error) {
	p := bytes.NewBuffer([]byte{})
	err := this.Cons.Serialize(p)
	if err != nil {
		return nil, errors.NewDetailErr(err, errors.ErrNetPackFail, fmt.Sprintf("serialize error. consensus:%v", this.Cons))
	}
	return p.Bytes(), nil
}

func (this *Consensus) CmdType() string {
	return common.CONSENSUS_TYPE
}

//Deserialize message payload
func (this *Consensus) Deserialization(p []byte) error {
	log.Debug()
	buf := bytes.NewBuffer(p)
	err := this.Cons.Deserialize(buf)
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNetUnPackFail, fmt.Sprintf("deserialize Cons error. buf:%v", buf))
	}
	return nil
}
