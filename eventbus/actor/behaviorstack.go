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

package actor

type behaviorStack []ActorFunc

func (b *behaviorStack) Clear() {
	if len(*b) == 0 {
		return
	}

	for i := range *b {
		(*b)[i] = nil
	}
	*b = (*b)[:0]
}

func (b *behaviorStack) Peek() (v ActorFunc, ok bool) {
	l := b.Len()
	if l > 0 {
		ok = true
		v = (*b)[l-1]
	}
	return
}

func (b *behaviorStack) Push(v ActorFunc) {
	*b = append(*b, v)
}

func (b *behaviorStack) Pop() (v ActorFunc, ok bool) {
	l := b.Len()
	if l > 0 {
		l--
		ok = true
		v = (*b)[l]
		(*b)[l] = nil
		*b = (*b)[:l]
	}
	return
}

func (b *behaviorStack) Len() int {
	return len(*b)
}
