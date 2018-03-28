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

// Copyright 2017 The go-interpreter Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package stack implements a growable uint64 stack
package stack

type Stack struct {
	slice []uint64
}

func (s *Stack) Push(b uint64) {
	s.slice = append(s.slice, b)
}

func (s *Stack) Pop() uint64 {
	v := s.Top()
	s.slice = s.slice[:len(s.slice)-1]
	return v
}

func (s *Stack) SetTop(v uint64) {
	s.slice[len(s.slice)-1] = v
}

func (s *Stack) Top() uint64 {
	return s.slice[len(s.slice)-1]
}

func (s *Stack) Get(i int) uint64 {
	return s.slice[i]
}

func (s *Stack) Set(i int, v uint64) {
	s.slice[i] = v
}

func (s *Stack) Len() int {
	return len(s.slice)
}
