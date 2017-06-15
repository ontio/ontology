package sm2

import (
	"DNA/crypto/sm3"
	"DNA/crypto/util"
	"crypto/aes"
	"crypto/cipher"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha512"
	"errors"
	"io"
	"math/big"
)

type zr struct {
	io.Reader
}

const (
	aesIV  = "IV for <SM2> CTR"
	USERID = "1234567812345678"
)

var paramA *big.Int

var zeroReader = &zr{}
var p256_sm2 *elliptic.CurveParams
var one = new(big.Int).SetInt64(1)

type combinedMult interface {
	CombinedMult(bigX, bigY *big.Int, baseScalar, scalar []byte) (x, y *big.Int)
}

func (z *zr) Read(dst []byte) (n int, err error) {
	for i := range dst {
		dst[i] = 0
	}
	return len(dst), nil
}

func Init(algSet *util.CryptoAlgSet) {
	p256_sm2 = &elliptic.CurveParams{Name: "SM2-P-256"}
	p256_sm2.P, _ = new(big.Int).SetString("FFFFFFFEFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFF00000000FFFFFFFFFFFFFFFF", 16)
	p256_sm2.N, _ = new(big.Int).SetString("FFFFFFFEFFFFFFFFFFFFFFFFFFFFFFFF7203DF6B21C6052B53BBF40939D54123", 16)
	p256_sm2.B, _ = new(big.Int).SetString("28E9FA9E9D9F5E344D5A9E4BCF6509A7F39789F515AB8F92DDBCBD414D940E93", 16)
	p256_sm2.Gx, _ = new(big.Int).SetString("32C4AE2C1F1981195F9904466A39C9948FE30BBFF2660BE1715A4589334C74C7", 16)
	p256_sm2.Gy, _ = new(big.Int).SetString("BC3736A2F4F6779C59BDCEE36B692153D0A9877CC62A474002DF32E52139F0A0", 16)
	p256_sm2.BitSize = 256

	paramA, _ = new(big.Int).SetString("FFFFFFFEFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFF00000000FFFFFFFFFFFFFFFC", 16)

	algSet.EccParams = *p256_sm2
	algSet.Curve = p256_sm2
}

func randFieldElement(c elliptic.Curve, rand io.Reader) (*big.Int, error) {
	params := c.Params()
	b := make([]byte, params.BitSize/8+8)
	_, err := io.ReadFull(rand, b)
	if err != nil {
		return nil, err
	}

	k := new(big.Int).SetBytes(b)
	n := new(big.Int).Sub(params.N, one)
	n = n.Sub(n, one) //n-2

	// 1 <= k <= n-2
	k.Mod(k, n)
	k.Add(k, one)
	return k, nil
}

func GenKeyPair(algSet *util.CryptoAlgSet) ([]byte, *big.Int, *big.Int, error) {
	k, err := randFieldElement(algSet.Curve, rand.Reader)
	if err != nil {
		return nil, nil, nil, errors.New("Generate key pair error")
	}

	publicKeyX, publicKeyY := algSet.Curve.ScalarBaseMult(k.Bytes())
	return k.Bytes(), publicKeyX, publicKeyY, nil
}

func Sign(algSet *util.CryptoAlgSet, priKey []byte, data []byte) (r *big.Int, s *big.Int, err error) {
	publicKeyX, publicKeyY := algSet.Curve.ScalarBaseMult(priKey)
	hash := sm3.Sum(combineZ(data, publicKeyX, publicKeyY, USERID))
	entropyLen := (algSet.EccParams.BitSize + 7) / 16
	if entropyLen > 32 {
		entropyLen = 32
	}
	entropy := make([]byte, entropyLen)
	_, err = io.ReadFull(rand.Reader, entropy)
	if err != nil {
		return nil, nil, err
	}

	md := sha512.New()
	md.Write(priKey)
	md.Write(entropy)
	md.Write(hash[:])
	key := md.Sum(nil)[:32]

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, nil, err
	}

	cspRng := cipher.StreamReader{
		R: zeroReader,
		S: cipher.NewCTR(block, []byte(aesIV)),
	}

	c := algSet.Curve
	N := c.Params().N
	if N.Sign() == 0 {
		return nil, nil, errors.New("zero parameter")
	}
	var k *big.Int
	e := new(big.Int).SetBytes(hash[:])
	for {
		for {
			k, err = randFieldElement(c, cspRng)
			if err != nil {
				r = nil
				return nil, nil, errors.New("randFieldElement error")
			}

			r, _ = algSet.Curve.ScalarBaseMult(k.Bytes())
			r.Add(r, e)
			r.Mod(r, N)
			if r.Sign() != 0 {
				break
			}
			if t := new(big.Int).Add(r, k); t.Cmp(N) == 0 {
				break
			}
		}
		D := new(big.Int).SetBytes(priKey)
		rD := new(big.Int).Mul(D, r)
		s = new(big.Int).Sub(k, rD)
		d1 := new(big.Int).Add(D, one)
		d1Inv := new(big.Int).ModInverse(d1, N)
		s.Mul(s, d1Inv)
		s.Mod(s, N)
		if s.Sign() != 0 {
			break
		}
	}

	return
}

func Verify(algSet *util.CryptoAlgSet, publicKeyX *big.Int, publicKeyY *big.Int, data []byte, r, s *big.Int) (bool, error) {
	c := algSet.Curve
	N := c.Params().N

	if r.Sign() <= 0 || s.Sign() <= 0 {
		return false, errors.New("SM2 signature contained zero or negative values")
	}
	if r.Cmp(N) >= 0 || s.Cmp(N) >= 0 {
		return false, errors.New("SM2 signature contained zero or negative values")
	}

	t := new(big.Int).Add(r, s)
	t.Mod(t, N)
	if N.Sign() == 0 {
		return false, errors.New("SM2 Params N contained zero or negative values")
	}

	var x *big.Int
	if opt, ok := c.(combinedMult); ok {
		x, _ = opt.CombinedMult(publicKeyX, publicKeyY, s.Bytes(), t.Bytes())
	} else {
		x1, y1 := c.ScalarBaseMult(s.Bytes())
		x2, y2 := c.ScalarMult(publicKeyX, publicKeyY, t.Bytes())
		x, _ = c.Add(x1, y1, x2, y2)
	}

	hash := sm3.Sum(combineZ(data, publicKeyX, publicKeyY, USERID))
	e := new(big.Int).SetBytes(hash[:])
	x.Add(x, e)
	x.Mod(x, N)
	return x.Cmp(r) == 0, nil
}

// Combine the raw data with user ID, curve parameters and public key
// to generate the signed data used in Sign and Verify
func combineZ(raw []byte, publicKeyX, publicKeyY *big.Int, userID string) []byte {
	h := sm3.New()

	id := []byte(userID)
	len := len(id) * 8
	blen := []byte{byte((len >> 8) & 0xff), byte(len & 0xff)}

	h.Write(blen)
	h.Write(id)
	h.Write(paramA.Bytes())
	h.Write(p256_sm2.B.Bytes())
	h.Write(p256_sm2.Gx.Bytes())
	h.Write(p256_sm2.Gy.Bytes())
	h.Write(publicKeyX.Bytes())
	h.Write(publicKeyY.Bytes())
	return append(h.Sum(make([]byte, 0)), raw...)
}
