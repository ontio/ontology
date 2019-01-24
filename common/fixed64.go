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

package common

import (
	"bytes"
	"io"
	"strconv"
)

//the 64 bit fixed-point number, precise 10^-8
type Fixed64 int64

const (
	Decimal = 100000000
)

func (f *Fixed64) Serialization(sink *ZeroCopySink) {
	sink.WriteInt64(int64(*f))
}

func (f *Fixed64) Deserialization(source *ZeroCopySource) error {
	var eof bool

	t, eof := source.NextInt64()
	if eof {
		return io.ErrUnexpectedEOF
	}
	*f = Fixed64(t)
	return nil
}

func FromDecimal(value int64) Fixed64 {
	return Fixed64(value * Decimal)
}

func (f Fixed64) GetData() int64 {
	return int64(f)
}

func (f Fixed64) String() string {
	var buffer bytes.Buffer
	value := int64(f)
	if value < 0 {
		buffer.WriteRune('-')
		value = -value
	}
	buffer.WriteString(strconv.FormatInt(value/100000000, 10))
	value %= 100000000
	if value > 0 {
		buffer.WriteRune('.')
		s := strconv.FormatInt(value, 10)
		for i := len(s); i < 8; i++ {
			buffer.WriteRune('0')
		}
		buffer.WriteString(s)
	}
	return buffer.String()
}
