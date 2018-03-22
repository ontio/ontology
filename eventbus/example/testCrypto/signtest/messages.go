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

package signtest

import "github.com/Ontology/crypto"

type SignRequest struct {
	Data []byte
	Seq  string
}

type SignResponse struct {
	Signature []byte
	Seq  string
}

type SetPrivKey struct{
	PrivKey []byte
}



type VerifyRequest struct {
	Signature []byte
	Data []byte
	PublicKey crypto.PubKey
	Seq string
}

type VerifyResponse struct {
	Seq string
	Result bool
	ErrorMsg string
}