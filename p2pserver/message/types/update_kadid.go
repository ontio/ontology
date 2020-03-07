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
	common2 "github.com/ontio/ontology/common"
	"github.com/ontio/ontology/p2pserver/common"
	"github.com/ontio/ontology/p2pserver/dht/kbucket"
)

type UpdateKadId struct {
	//TODO remove this legecy field when upgrade network layer protocal
	KadKeyId *kbucket.KadKeyId
}

//Serialize message payload
func (this *UpdateKadId) Serialization(sink *common2.ZeroCopySink) {
	this.KadKeyId.Serialization(sink)
}

func (this *UpdateKadId) Deserialization(source *common2.ZeroCopySource) error {
	this.KadKeyId = &kbucket.KadKeyId{}
	return this.KadKeyId.Deserialization(source)
}

func (this *UpdateKadId) CmdType() string {
	return common.UPDATE_KADID_TYPE
}
