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

package dbft

type ConsensusState byte

const (
	Initial         ConsensusState = 0x00
	Primary         ConsensusState = 0x01
	Backup          ConsensusState = 0x02
	RequestSent     ConsensusState = 0x04
	RequestReceived ConsensusState = 0x08
	SignatureSent   ConsensusState = 0x10
	BlockGenerated  ConsensusState = 0x20
)

func (state ConsensusState) HasFlag(flag ConsensusState) bool {
	return (state & flag) == flag
}
