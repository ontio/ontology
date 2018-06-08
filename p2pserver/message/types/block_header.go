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
	"encoding/binary"
	"fmt"

	"github.com/ontio/ontology/common/serialization"
	ct "github.com/ontio/ontology/core/types"
	"github.com/ontio/ontology/errors"
	"github.com/ontio/ontology/p2pserver/common"
)

type BlkHeader struct {
	BlkHdr []*ct.Header
}

//Serialize message payload
func (this BlkHeader) Serialization() ([]byte, error) {
	p := bytes.NewBuffer([]byte{})
	serialization.WriteUint32(p, uint32(len(this.BlkHdr)))
	for _, header := range this.BlkHdr {
		err := header.Serialize(p)
		if err != nil {
			return nil, errors.NewDetailErr(err, errors.ErrNetPackFail, fmt.Sprintf("serialize error. header:%v", header))
		}
	}

	return p.Bytes(), nil
}

func (this *BlkHeader) CmdType() string {
	return common.HEADERS_TYPE
}

//Deserialize message payload
func (this *BlkHeader) Deserialization(p []byte) error {
	buf := bytes.NewBuffer(p)
	var count uint32

	err := binary.Read(buf, binary.LittleEndian, &count)
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNetUnPackFail, fmt.Sprintf("read Cnt error. buf:%v", buf))
	}

	for i := 0; i < int(count); i++ {
		var headers ct.Header
		err := headers.Deserialize(buf)
		if err != nil {
			return errors.NewDetailErr(err, errors.ErrNetUnPackFail, fmt.Sprintf("deserialize headers error. buf:%v", buf))
		}
		this.BlkHdr = append(this.BlkHdr, &headers)
	}
	return nil
}
