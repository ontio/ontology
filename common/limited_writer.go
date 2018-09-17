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
	"errors"
	"io"
)

var ErrWriteExceedLimitedCount = errors.New("writer exceed limited count")

type LimitedWriter struct {
	count  uint64
	max    uint64
	writer io.Writer
}

func NewLimitedWriter(w io.Writer, max uint64) *LimitedWriter {
	return &LimitedWriter{
		writer: w,
		max:    max,
	}
}

func (self *LimitedWriter) Write(buf []byte) (int, error) {
	if self.count+uint64(len(buf)) > self.max {
		return 0, ErrWriteExceedLimitedCount
	}
	n, err := self.writer.Write(buf)
	self.count += uint64(n)
	return n, err
}

// Count function return counted bytes
func (self *LimitedWriter) Count() uint64 {
	return self.count
}
