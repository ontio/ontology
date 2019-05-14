/*
 * Copyright (C) 2019 The DAD Authors
 * This file is part of The DAD library.
 *
 * The DAD is free software: you can redistribute it and/or modify
 * it under the terms of the GNU Lesser General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * The DAD is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU Lesser General Public License for more details.
 *
 * You should have received a copy of the GNU Lesser General Public License
 * along with The DAD.  If not, see <http://www.gnu.org/licenses/>.
 */

// Package gaillier This package implements a Paillier cryptosystem
package gaillier

import (
	"crypto/rand"
	"errors"
	"io"
	"math/big"
)

var one = big.NewInt(1)

// ErrMessageTooLong is returned when attempting to encrypt a message which is
// too large for the size of the public key.
var ErrMessageTooLong = errors.New("paillier: message too long for Paillier public key size")

// GenerateKey generates an Paillier keypair of the given bit size using the
// random source random (for example, crypto/rand.Reader).
func GenerateKey(random io.Reader, bits int) (*PrivateKey, error) {
	// First, begin generation of p in the background.
	var p *big.Int
	var errChan = make(chan error, 1)
	go func() {
		var err error
		p, err = rand.Prime(random, bits/2)
		errChan <- err
	}()

	// Now, find a prime q in the foreground.
	q, err := rand.Prime(random, bits/2)
	if err != nil {
		return nil, err
	}

	// Wait for generation of p to complete successfully.
	if err := <-errChan; err != nil {
		return nil, err
	}

	n := new(big.Int).Mul(p, q)
	pp := new(big.Int).Mul(p, p)
	qq := new(big.Int).Mul(q, q)

	return &PrivateKey{
		PublicKey: PublicKey{
			N:        n,
			NSquared: new(big.Int).Mul(n, n),
			G:        new(big.Int).Add(n, one), // g = n + 1
		},
		p:         p,
		pp:        pp,
		pminusone: new(big.Int).Sub(p, one),
		q:         q,
		qq:        qq,
		qminusone: new(big.Int).Sub(q, one),
		pinvq:     new(big.Int).ModInverse(p, q),
		hp:        h(p, pp, n),
		hq:        h(q, qq, n),
		n:         n,
	}, nil

}

// PrivateKey represents a Paillier key.
type PrivateKey struct {
	PublicKey
	p         *big.Int
	pp        *big.Int
	pminusone *big.Int
	q         *big.Int
	qq        *big.Int
	qminusone *big.Int
	pinvq     *big.Int
	hp        *big.Int
	hq        *big.Int
	n         *big.Int
}

// PublicKey represents the public part of a Paillier key.
type PublicKey struct {
	N        *big.Int // modulus
	G        *big.Int // n+1, since p and q are same length
	NSquared *big.Int
}

func h(p *big.Int, pp *big.Int, n *big.Int) *big.Int {
	gp := new(big.Int).Mod(new(big.Int).Sub(one, n), pp)
	lp := l(gp, p)
	hp := new(big.Int).ModInverse(lp, p)
	return hp
}

func l(u *big.Int, n *big.Int) *big.Int {
	return new(big.Int).Div(new(big.Int).Sub(u, one), n)
}

// Encrypt encrypts a plain text represented as a byte array. The passed plain
// text MUST NOT be larger than the modulus of the passed public key.
func Encrypt(pubKey *PublicKey, plainText []byte) ([]byte, error) {
	c, _, err := EncryptAndNonce(pubKey, plainText)
	return c, err
}

// EncryptAndNonce encrypts a plain text represented as a byte array, and in
// addition, returns the nonce used during encryption. The passed plain text
// MUST NOT be larger than the modulus of the passed public key.
func EncryptAndNonce(pubKey *PublicKey, plainText []byte) ([]byte, *big.Int, error) {
	r, err := rand.Int(rand.Reader, pubKey.N)
	if err != nil {
		return nil, nil, err
	}

	c, err := EncryptWithNonce(pubKey, r, plainText)
	if err != nil {
		return nil, nil, err
	}

	return c.Bytes(), r, nil
}

// EncryptWithNonce encrypts a plain text represented as a byte array using the
// provided nonce to perform encryption. The passed plain text MUST NOT be
// larger than the modulus of the passed public key.
func EncryptWithNonce(pubKey *PublicKey, r *big.Int, plainText []byte) (*big.Int, error) {
	m := new(big.Int).SetBytes(plainText)
	if pubKey.N.Cmp(m) < 1 { // N < m
		return nil, ErrMessageTooLong
	}

	// c = g^m * r^n mod n^2 = ((m*n+1) mod n^2) * r^n mod n^2
	n := pubKey.N
	c := new(big.Int).Mod(
		new(big.Int).Mul(
			new(big.Int).Mod(new(big.Int).Add(one, new(big.Int).Mul(m, n)), pubKey.NSquared),
			new(big.Int).Exp(r, n, pubKey.NSquared),
		),
		pubKey.NSquared,
	)

	return c, nil
}

// Decrypt decrypts the passed cipher text.
func Decrypt(privKey *PrivateKey, cipherText []byte) ([]byte, error) {
	c := new(big.Int).SetBytes(cipherText)
	if privKey.NSquared.Cmp(c) < 1 { // c < n^2
		return nil, ErrMessageTooLong
	}

	cp := new(big.Int).Exp(c, privKey.pminusone, privKey.pp)
	lp := l(cp, privKey.p)
	mp := new(big.Int).Mod(new(big.Int).Mul(lp, privKey.hp), privKey.p)
	cq := new(big.Int).Exp(c, privKey.qminusone, privKey.qq)
	lq := l(cq, privKey.q)

	mqq := new(big.Int).Mul(lq, privKey.hq)
	mq := new(big.Int).Mod(mqq, privKey.q)
	m := crt(mp, mq, privKey)

	return m.Bytes(), nil
}

func crt(mp *big.Int, mq *big.Int, privKey *PrivateKey) *big.Int {
	u := new(big.Int).Mod(new(big.Int).Mul(new(big.Int).Sub(mq, mp), privKey.pinvq), privKey.q)
	m := new(big.Int).Add(mp, new(big.Int).Mul(u, privKey.p))
	return new(big.Int).Mod(m, privKey.n)
}
