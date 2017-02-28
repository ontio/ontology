package sm2

import (
	"crypto/elliptic"
	"fmt"
	"math/big"
)

// ECPoint ---
type ECPoint struct {
	X, Y  *ECFieldElement
	curve *elliptic.CurveParams
}

// PrintHex ---
func PrintHex(str string, bt []byte, length int) {
	fmt.Println(str, "Length = ", length)
	for i := 0; i < length; i++ {
		if i%16 == 0 && i != 0 {
			fmt.Println()
		}
		fmt.Printf("0x%02x ", bt[i])
	}
	fmt.Println(" ")
	fmt.Println(" ")
}

// NewECPoint ---
func NewECPoint() *ECPoint {
	x := NewECFieldElement()
	y := NewECFieldElement()
	return &ECPoint{x, y, EcParams}
}

// DumpECPoint ---
func DumpECPoint(dst *ECPoint, src *ECPoint) {
	DumpECFieldElement(dst.X, src.X)
	DumpECFieldElement(dst.Y, src.Y)
	dst.curve = src.curve
}

// IsInfinity ---
func (e *ECPoint) IsInfinity() bool {
	if e.X == nil && e.Y == nil {
		return true
	}

	if e.X.value.Cmp(big.NewInt(0)) == 0 && e.Y.value.Cmp(big.NewInt(0)) == 0 {
		return true
	}
	return false
}

// Twice ---
func (e *ECPoint) Twice() *ECPoint {
	if e.IsInfinity() {
		return e
	}
	if 0 == e.Y.value.Sign() {
		return e
	}

	field2 := &ECFieldElement{big.NewInt(2), e.curve}
	field3 := &ECFieldElement{big.NewInt(3), e.curve}

	tmp := big.NewInt(0)
	tmp.Exp(e.X.value, big.NewInt(2), big.NewInt(0))

	gamma := NewECFieldElement()
	gamma.MulBig(field3, tmp)
	//gamma.AddBig(gamma, e.curve.A)
	gamma.AddBig(gamma, Sm2ParamA)

	fieldTmp := NewECFieldElement()
	fieldTmp.MulBig(field2, e.Y.value)
	gamma.Div(gamma, fieldTmp)
	gammaSqr := gamma.Square()

	fieldTmp.MulBig(field2, e.X.value)
	x := gammaSqr.Sub(gammaSqr, fieldTmp)

	y := NewECFieldElement()
	y.Sub(e.X, x)
	y.Mul(y, gamma)
	y.Sub(y, e.Y)

	return &ECPoint{x, y, e.curve}
}

// WindowNaf ---
func WindowNaf(width byte, k *big.Int) []int8 {
	wnaf := make([]int8, k.BitLen()+1)

	pow2wB := uint16(1 << width)

	length := 0
	bigp2wB := big.NewInt(int64(pow2wB))
	tmp := big.NewInt(0)

	for i := 0; k.Sign() > 0; i++ {
		if !isEven(k) {
			remainder := big.NewInt(0)
			remainder.Mod(k, bigp2wB)
			if remainder.Bit(int(width-1)) == 1 {
				tmp.Sub(remainder, bigp2wB)
				wnaf[i] = int8(tmp.Int64())
			} else {
				wnaf[i] = int8(remainder.Int64())
			}

			k.Sub(k, big.NewInt(int64(wnaf[i])))
			length = i
		} else {
			wnaf[i] = 0
		}
		k.Rsh(k, 1)
	}
	length++
	wnafShort := make([]int8, length)
	copy(wnafShort, wnaf[0:length])
	return wnafShort
}

// Multiply ---
func Multiply(p *ECPoint, k *big.Int) *ECPoint {
	m := k.BitLen()

	var width byte
	var reqPreCompLen int

	if m < 13 {
		width = 2
		reqPreCompLen = 1
	} else if m < 41 {
		width = 3
		reqPreCompLen = 2
	} else if m < 121 {
		width = 4
		reqPreCompLen = 4
	} else if m < 337 {
		width = 5
		reqPreCompLen = 8
	} else if m < 897 {
		width = 6
		reqPreCompLen = 16
	} else if m < 2305 {
		width = 7
		reqPreCompLen = 32
	} else {
		width = 8
		reqPreCompLen = 127
	}

	preCompLen := 1
	preComp := make([]*ECPoint, reqPreCompLen)
	for i := 0; i < reqPreCompLen; i++ {
		preComp[i] = NewECPoint()
	}

	DumpECPoint(preComp[0], p)

	twiceP := p.Twice()

	if preCompLen < reqPreCompLen {
		oldPreComp := NewECPoint()
		DumpECPoint(oldPreComp, preComp[0])
		DumpECPoint(preComp[0], oldPreComp)

		for i := preCompLen; i < reqPreCompLen; i++ {
			preComp[i].Add(twiceP, preComp[i-1])
		}
	}

	wnaf := WindowNaf(width, k)
	l := len(wnaf)

	product := NewECPoint()
	for i := l - 1; i >= 0; i-- {
		product = product.Twice()

		id := wnaf[i]
		if 0 != id {
			if id > 0 {
				index := (id - 1) / 2
				product.Add(product, preComp[index])
			} else {
				index := (-id - 1) / 2
				product.Sub(product, preComp[index])
			}
		}
	}
	return product
}

// Neg ---
func (e *ECPoint) Neg(x *ECPoint) *ECPoint {
	negY := NewECFieldElement()
	negY.Neg(x.Y)
	negPoint := &ECPoint{x.X, negY, x.curve}

	DumpECPoint(e, negPoint)

	return negPoint
}

// Mul ---
func (e *ECPoint) Mul(p *ECPoint, n []byte) *ECPoint {
	if nil == p || nil == n {
		fmt.Println("Argument is Null")
	}
	if len(n) > 32 {
		fmt.Println("Argument is not excepted", len(n))
	}
	if p.IsInfinity() {
		return p
	}

	k := new(big.Int).SetBytes(n)
	if 0 == k.Sign() {
		return Infinity
	}
	nP := Multiply(p, k)
	DumpECPoint(e, nP)
	return nP
}

// Add ---
func (e *ECPoint) Add(x *ECPoint, y *ECPoint) *ECPoint {
	if x.IsInfinity() {
		DumpECPoint(e, y)
		return y
	}
	if y.IsInfinity() {
		DumpECPoint(e, y)
		return x
	}
	if 0 == x.X.value.Cmp(y.X.value) {
		if 0 == x.Y.value.Cmp(y.Y.value) {
			twcX := x.Twice()
			DumpECPoint(e, twcX)
			return twcX
		}
		return Infinity
	}

	field1 := NewECFieldElement()
	field1.Sub(y.Y, x.Y)

	field2 := NewECFieldElement()
	field2.Sub(y.X, x.X)

	gama := NewECFieldElement()
	gama.Div(field1, field2)

	x3 := gama.Square()
	x3.Sub(x3, x.X)
	x3.Sub(x3, y.X)

	y3 := NewECFieldElement()
	y3.Sub(x.X, x3)
	y3.Mul(y3, gama)
	y3.Sub(y3, x.Y)

	ret := &ECPoint{x3, y3, x.curve}
	DumpECPoint(e, ret)
	return ret

}

// Sub ---
func (e *ECPoint) Sub(x *ECPoint, y *ECPoint) *ECPoint {
	if y.IsInfinity() {
		return x
	}

	point3 := NewECPoint()
	point3.Neg(y)
	point3.Add(point3, x)
	DumpECPoint(e, point3)
	return point3
}

// SumOfTwoMultiplies ---
func SumOfTwoMultiplies(P *ECPoint, k *big.Int, Q *ECPoint, l *big.Int) *ECPoint {
	m := 0
	if k.BitLen() > l.BitLen() {
		m = k.BitLen()
	} else {
		m = l.BitLen()
	}

	sumPQ := NewECPoint()
	sumPQ.Add(P, Q)

	sum := NewECPoint()
	DumpECPoint(sum, Infinity)

	for i := m - 1; i >= 0; i-- {
		sum = sum.Twice()

		if k.Bit(int(i)) == 1 {
			if l.Bit(int(i)) == 1 {
				sum.Add(sum, sumPQ)
			} else {
				sum.Add(sum, P)
			}
		} else {
			if l.Bit(int(i)) == 1 {
				sum.Add(sum, Q)
			}
		}
	}
	return sum
}
