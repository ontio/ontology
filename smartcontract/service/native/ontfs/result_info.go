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

package ontfs

import (
	"fmt"

	"github.com/ontio/ontology/common"
)

type RetInfo struct {
	Ret  bool
	Info []byte
}

func (this *RetInfo) Serialization(sink *common.ZeroCopySink) {
	sink.WriteBool(this.Ret)
	sink.WriteVarBytes(this.Info)
}

func (this *RetInfo) Deserialization(source *common.ZeroCopySource) error {
	var err error
	if this.Ret, err = DecodeBool(source); err != nil {
		return fmt.Errorf("[RetInfo] [Ret] Deserialization from error:%v", err)
	}
	if this.Info, err = DecodeVarBytes(source); err != nil {
		return fmt.Errorf("[RetInfo] [Info] Deserialization from error:%v", err)
	}
	return nil
}

func EncRet(ret bool, info []byte) []byte {
	retInfo := RetInfo{ret, info}
	sink := common.NewZeroCopySink(nil)
	retInfo.Serialization(sink)
	return sink.Bytes()
}

func DecRet(ret []byte) *RetInfo {
	var retInfo RetInfo
	source := common.NewZeroCopySource(ret)
	retInfo.Deserialization(source)
	return &retInfo
}
