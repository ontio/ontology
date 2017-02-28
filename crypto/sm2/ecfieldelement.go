package sm2

import (
	"crypto/elliptic"
	"crypto/rand"
	"fmt"
	"math/big"
)

type ECFieldElement struct {
	value      *big.Int
	curveParam *elliptic.CurveParams
}

func NewECFieldElement() *ECFieldElement {
	fieldValue := big.NewInt(0)
	return &ECFieldElement{fieldValue, EcParams}
}

func DumpECFieldElement(dst *ECFieldElement, src *ECFieldElement) {
	if nil != dst && nil != src {
		dst.value.Set(src.value)
		dst.curveParam = src.curveParam
	}
}

func isEven(bn *big.Int) bool {
	v := big.NewInt(0)
	v.Mod(bn, big.NewInt(2))
	if 0 == v.Int64() {
		return true
	}
	return false
}

func getLowestSetBit(k *big.Int) int {
	i := 0
	for i = 0; k.Bit(i) != 1; i++ {
	}
	return i
}

// fastLucasSequence refer to https://en.wikipedia.org/wiki/Lucas_sequence
func fastLucasSequence(curveP, lucasParamP, lucasParamQ, k *big.Int) (*big.Int,
	*big.Int) {
	n := k.BitLen()
	s := getLowestSetBit(k)

	uh := big.NewInt(1)
	vl := big.NewInt(2)
	ql := big.NewInt(1)
	qh := big.NewInt(1)
	vh := big.NewInt(0).Set(lucasParamP)
	tmp := big.NewInt(0)

	for j := n - 1; j >= s+1; j-- {
		ql.Mul(ql, qh)
		ql.Mod(ql, curveP)

		if k.Bit(j) == 1 {
			qh.Mul(ql, lucasParamQ)
			qh.Mod(qh, curveP)

			uh.Mul(uh, vh)
			uh.Mod(uh, curveP)

			vl.Mul(vh, vl)
			tmp.Mul(lucasParamP, ql)
			vl.Sub(vl, tmp)
			vl.Mod(vl, curveP)

			vh.Mul(vh, vh)
			tmp.Lsh(qh, 1)
			vh.Sub(vh, tmp)
			vh.Mod(vh, curveP)

		} else {
			qh.Set(ql)

			uh.Mul(uh, vl)
			uh.Sub(uh, ql)
			uh.Mod(uh, curveP)

			vh.Mul(vh, vl)
			tmp.Mul(lucasParamP, ql)
			vh.Sub(vh, tmp)
			vh.Mod(vh, curveP)

			vl.Mul(vl, vl)
			tmp.Lsh(ql, 1)
			vl.Sub(vl, tmp)
			vl.Mod(vl, curveP)
		}
	}

	ql.Mul(ql, qh)
	ql.Mod(ql, curveP)

	qh.Mul(ql, lucasParamQ)
	qh.Mod(qh, curveP)

	uh.Mul(uh, vl)
	uh.Sub(uh, ql)
	uh.Mod(uh, curveP)

	vl.Mul(vh, vl)
	tmp.Mul(lucasParamP, ql)
	vl.Sub(vl, tmp)
	vl.Mod(vl, curveP)

	ql.Mul(ql, qh)
	ql.Mod(ql, curveP)

	for j := 1; j <= s; j++ {
		uh.Mul(uh, vl)
		uh.Mul(uh, curveP)

		vl.Mul(vl, vl)
		tmp.Lsh(ql, 1)
		vl.Sub(vl, tmp)
		vl.Mod(vl, curveP)

		ql.Mul(ql, ql)
		ql.Mod(ql, curveP)
	}

	return uh, vl
}

// Square ---
func (e *ECFieldElement) Square() *ECFieldElement {
	eSquare := big.NewInt(0)
	eSquare.Mul(e.value, e.value)
	eSquare.Mod(eSquare, e.curveParam.P)

	return &ECFieldElement{eSquare, e.curveParam}
}

// Sqrt - compute the coordinate of Y from Y**2
func Sqrt(ySquare *big.Int, curve *elliptic.CurveParams) *big.Int {
	if curve.P.Bit(1) == 1 {
		tmp1 := big.NewInt(0)
		tmp1.Rsh(curve.P, 2)
		tmp1.Add(tmp1, big.NewInt(1))

		tmp2 := big.NewInt(0)
		tmp2.Exp(ySquare, tmp1, curve.P)

		tmp3 := big.NewInt(0)
		tmp3.Exp(tmp2, big.NewInt(2), curve.P)

		if 0 == tmp3.Cmp(ySquare) {
			return tmp2
		}
		fmt.Println("error z^2 != z")
		return nil
	}

	qMinusOne := big.NewInt(0)
	qMinusOne.Sub(curve.P, big.NewInt(1))

	legendExponent := big.NewInt(0)
	legendExponent.Rsh(qMinusOne, 1)

	tmp4 := big.NewInt(0)
	tmp4.Exp(ySquare, legendExponent, curve.P)
	if tmp4.Cmp(big.NewInt(1)) != 0 {
		return nil
	}

	k := big.NewInt(0)
	k.Rsh(qMinusOne, 2)
	k.Lsh(k, 1)
	k.Add(k, big.NewInt(1))

	lucasParamQ := big.NewInt(0)
	lucasParamQ.Set(ySquare)
	fourQ := big.NewInt(0)
	fourQ.Lsh(lucasParamQ, 2)
	fourQ.Mod(fourQ, curve.P)

	seqU := big.NewInt(0)
	seqV := big.NewInt(0)

	for {
		lucasParamP := big.NewInt(0)
		for {
			tmp5 := big.NewInt(0)
			lucasParamP, _ = rand.Prime(rand.Reader, curve.P.BitLen())

			if lucasParamP.Cmp(curve.P) < 0 {
				tmp5.Mul(lucasParamP, lucasParamP)
				tmp5.Sub(tmp5, fourQ)
				tmp5.Exp(tmp5, legendExponent, curve.P)
				if tmp5.Cmp(qMinusOne) == 0 {
					break
				}
			}
		}

		seqU, seqV = fastLucasSequence(curve.P, lucasParamP, lucasParamQ, k)

		tmp6 := big.NewInt(0)
		tmp6.Mul(seqV, seqV)
		tmp6.Mod(tmp6, curve.P)
		if tmp6.Cmp(fourQ) == 0 {
			if seqV.Bit(0) == 1 {
				seqV.Add(seqV, curve.P)
			}
			seqV.Rsh(seqV, 1)
			return seqV
		}
		if (seqU.Cmp(big.NewInt(1)) == 0) || (seqU.Cmp(qMinusOne) == 0) {
			break
		}
	}
	return nil
}

// ToByteArray --- NoUsed
func (e *ECFieldElement) ToByteArray() []byte {
	eData := e.value.Bytes()
	eLen := len(eData)

	byteArray := make([]byte, eLen)
	copy(byteArray, eData)

	return byteArray
}

// Mul ---
func (e *ECFieldElement) Mul(x *ECFieldElement, y *ECFieldElement) *ECFieldElement {
	return e.MulBig(x, y.value)
}

// Div ---
func (e *ECFieldElement) Div(x *ECFieldElement, y *ECFieldElement) *ECFieldElement {
	return e.DivBig(x, y.value)
}

// Add ---
func (e *ECFieldElement) Add(x *ECFieldElement, y *ECFieldElement) *ECFieldElement {
	return e.AddBig(x, y.value)
}

// Sub ---
func (e *ECFieldElement) Sub(x *ECFieldElement, y *ECFieldElement) *ECFieldElement {
	return e.SubBig(x, y.value)
}

// Neg ---
func (e *ECFieldElement) Neg(x *ECFieldElement) *ECFieldElement {

	negX := big.NewInt(0)
	negX.Neg(x.value)
	negX.Mod(negX, x.curveParam.P)

	e.value.Set(negX)

	return &ECFieldElement{negX, x.curveParam}
}

// MulBig ---
func (e *ECFieldElement) MulBig(x *ECFieldElement, y *big.Int) *ECFieldElement {
	pro := big.NewInt(0)
	pro.Mul(x.value, y)
	pro.Mod(pro, x.curveParam.P)

	e.value.Set(pro)

	return &ECFieldElement{pro, x.curveParam}
}

// DivBig ---
func (e *ECFieldElement) DivBig(x *ECFieldElement, y *big.Int) *ECFieldElement {
	inv := big.NewInt(0)
	inv.ModInverse(y, x.curveParam.P)

	quo := big.NewInt(0)
	quo.Mul(x.value, inv)
	quo.Mod(quo, x.curveParam.P)

	e.value.Set(quo)

	return &ECFieldElement{quo, x.curveParam}
}

// AddBig ---
func (e *ECFieldElement) AddBig(x *ECFieldElement, y *big.Int) *ECFieldElement {
	sum := big.NewInt(0)
	sum.Add(x.value, y)
	sum.Mod(sum, x.curveParam.P)

	e.value.Set(sum)

	return &ECFieldElement{sum, x.curveParam}
}

// SubBig ---
func (e *ECFieldElement) SubBig(x *ECFieldElement, y *big.Int) *ECFieldElement {

	dif := big.NewInt(0)
	dif.Sub(x.value, y)
	dif.Mod(dif, x.curveParam.P)

	e.value.Set(dif)

	return &ECFieldElement{dif, x.curveParam}
}
