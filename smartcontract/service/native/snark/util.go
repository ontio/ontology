package snark

import (
	"errors"
	"fmt"
	"math/big"
	"os"
	"strconv"

	"github.com/kunxian-xia/bn256"
)

var q, _ = new(big.Int).SetString("21888242871839275222246405745257275088696311157297823662689037894645226208583", 10)
var rinv, _ = new(big.Int).SetString("20988524275117001072002809824448087578619730785600314334253784976379291040311", 10)

func reverse(slice []byte) {
	for i, j := 0, len(slice)-1; i < j; i, j = i+1, j-1 {
		slice[i], slice[j] = slice[j], slice[i]
	}
}

// G1 point in the sprout-verifiying.key uses the Montgomery form and little-endian
func parseG1(f *os.File) (*bn256.G1, error) {
	isZero := make([]byte, 1)
	_, err := f.Read(isZero)
	if err != nil {
		return nil, err
	}

	if isZero[0] == 0x30 {
		raw := make([]byte, 64)
		_, err := f.Read(raw)
		if err != nil {
			return nil, err
		}

		reverse(raw[:32])
		reverse(raw[32:])

		xMont := new(big.Int).SetBytes(raw[:32])
		yMont := new(big.Int).SetBytes(raw[32:])

		x := xMont.Mul(xMont, rinv).Mod(xMont, q)
		y := yMont.Mul(yMont, rinv).Mod(yMont, q)

		copy(raw[:32], x.Bytes())
		copy(raw[32:], y.Bytes())

		g1 := new(bn256.G1)
		_, ok := g1.Unmarshal(raw)
		if !ok {
			return nil, errors.New("failed to unmarshal G1")
		}
		return g1, nil
	}
	return nil, errors.New("first byte != 0x30")
}

// G2 point in the sprout-verifiying.key uses the Montgomery form and little-endian
func parseG2(f *os.File) (*bn256.G2, error) {
	isZero := make([]byte, 1)
	_, err := f.Read(isZero)
	if err != nil {
		return nil, err
	}

	if isZero[0] == 0x30 {
		raw := make([]byte, 128)
		_, err := f.Read(raw)
		if err != nil {
			return nil, err
		}

		reverse(raw[:32])
		reverse(raw[32:64])
		reverse(raw[64:96])
		reverse(raw[96:])

		x0Mont := new(big.Int).SetBytes(raw[:32])
		x1Mont := new(big.Int).SetBytes(raw[32:64])
		y0Mont := new(big.Int).SetBytes(raw[64:96])
		y1Mont := new(big.Int).SetBytes(raw[96:])

		x0 := x0Mont.Mul(x0Mont, rinv).Mod(x0Mont, q)
		x1 := x1Mont.Mul(x1Mont, rinv).Mod(x1Mont, q)
		y0 := y0Mont.Mul(y0Mont, rinv).Mod(y0Mont, q)
		y1 := y1Mont.Mul(y1Mont, rinv).Mod(y1Mont, q)

		copy(raw[:32], x1.Bytes())
		copy(raw[32:64], x0.Bytes())
		copy(raw[64:96], y1.Bytes())
		copy(raw[96:], y0.Bytes())

		g2 := new(bn256.G2)
		g2, ok := g2.Unmarshal(raw)
		if !ok {
			return nil, errors.New("failed to unmarshal G2")
		}
		return g2, nil
	}
	return nil, errors.New("first byte != 0x30")
}

func parseVK(f *os.File) (*phgr13VerifyingKey, error) {
	var err error
	vk := new(phgr13VerifyingKey)
	vk.a, err = parseG2(f)
	if err != nil {
		return nil, err
	}

	vk.b, err = parseG1(f)
	if err != nil {
		return nil, err
	}

	vk.c, err = parseG2(f)
	if err != nil {
		return nil, err
	}

	vk.gamma, err = parseG2(f)
	if err != nil {
		return nil, err
	}

	vk.gammaBeta1, err = parseG1(f)
	if err != nil {
		return nil, err
	}

	vk.gammaBeta2, err = parseG2(f)
	if err != nil {
		return nil, err
	}

	vk.z, err = parseG2(f)
	if err != nil {
		return nil, err
	}

	vk.icLen = 10
	vk.ic = make([]*bn256.G1, vk.icLen)

	vk.ic[0], err = parseG1(f)
	if err != nil {
		return nil, err
	}
	f.Seek(12*2, 1)
	for i := 0; i < 9; i++ {
		vk.ic[i+1], err = parseG1(f)
		if err != nil {
			return nil, err
		}
	}
	return vk, nil
}

func parseCompressedG1(raw []byte) (*bn256.G1, error) {
	if len(raw) != 33 {
		return nil, errors.New("len(raw) != 33")
	}

	if raw[0]&0x02 != 0x02 {
		return nil, errors.New("wrong prefix " + strconv.Itoa(int(raw[0])))
	}

	raw[0] = raw[0] & 0x01
	g1 := new(bn256.G1)
	g1, ok := g1.Decompress(raw)
	if !ok {
		return nil, errors.New("failed to decompress")
	}
	return g1, nil
}

func importProof(raw []byte) (*phgr13Proof, error) {
	var err error
	var G1Size = 33
	var G2Size = 65
	proof := new(phgr13Proof)

	proof.a, err = parseCompressedG1(raw[:G1Size])
	if err != nil {
		fmt.Println("a")
		return nil, err
	}

	proof.aPrime, err = parseCompressedG1(raw[G1Size : 2*G1Size])
	if err != nil {
		fmt.Println("a'")
		return nil, err
	}

	//proof.PiB, err = parseCompressedG2(raw[2*G1Size : 2*G1Size+G2Size])

	proof.bPrime, err = parseCompressedG1(raw[2*G1Size+G2Size : 3*G1Size+G2Size])
	if err != nil {
		fmt.Println("b'")
		return nil, err
	}

	proof.c, err = parseCompressedG1(raw[3*G1Size+G2Size : 4*G1Size+G2Size])
	if err != nil {
		fmt.Println("c")
		return nil, err
	}

	proof.cPrime, err = parseCompressedG1(raw[4*G1Size+G2Size : 5*G1Size+G2Size])
	if err != nil {
		fmt.Println("c'")
		return nil, err
	}

	proof.k, err = parseCompressedG1(raw[5*G1Size+G2Size : 6*G1Size+G2Size])
	if err != nil {
		fmt.Println("k")
		return nil, err
	}

	proof.h, err = parseCompressedG1(raw[6*G1Size+G2Size : 7*G1Size+G2Size])
	if err != nil {
		fmt.Println("h")
		return nil, err
	}

	return proof, nil
}
