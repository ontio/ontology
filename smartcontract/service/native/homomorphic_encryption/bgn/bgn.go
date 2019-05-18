package bgn

import (
	"crypto/rand"
	"log"
	"math"
	"math/big"
	"strconv"
	"strings"

	"github.com/Nik-U/pbc"
)

// PublicKey is the BGN public key used for encryption
// as well as performing homomorphic operations on ciphertexts
type PublicKey struct {
	Pairing       *pbc.Pairing // pairing between G1 and G2
	G1            *pbc.Element // G1 group
	P             *pbc.Element // generator of G1
	Q             *pbc.Element
	N             *big.Int // product of two primes
	T             *big.Int // message space T
	PolyBase      int      // ciphertext polynomial encoding base
	FPScaleBase   int      // fixed point encoding scale base
	FPPrecision   float64  // min error tolerance for fixed point encoding
	Deterministic bool     // whether or not the homomorphic operations are deterministic
}

// SecretKey used for decryption of ciphertexts
type SecretKey struct {
	Key      *big.Int
	PolyBase int
}

// NewKeyGen creates a new public/private key pair of size bits
func NewKeyGen(keyBits int, T *big.Int, polyBase int, fpScaleBase int, fpPrecision float64, deterministic bool) (*PublicKey, *SecretKey, error) {

	if keyBits < 16 {
		panic("key bits must be >= 16 bits in length")
	}

	var q1 *big.Int    // random prime
	var q2 *big.Int    // secret key (random prime)
	var N *big.Int     // n = r*q
	var P *pbc.Element // field element
	var Q *pbc.Element // field element

	// generate a new random prime r
	q1, err := rand.Prime(rand.Reader, keyBits)

	// generate a new random prime q (this will be the secret key)
	q2, err = rand.Prime(rand.Reader, keyBits)

	if err != nil {
		return nil, nil, err
	}

	if q1.Cmp(T) < 0 || q2.Cmp(T) < 0 {
		panic("Message space is greater than the group order!")
	}

	// compute the product of the primes
	N = big.NewInt(0).Mul(q1, q2)
	params := pbc.GenerateA1(N)

	if err != nil {
		return nil, nil, err
	}

	// create a new pairing with given params
	pairing := pbc.NewPairing(params)

	// generate the two multiplicative groups of
	// order n (using pbc pairing library)
	G1 := pairing.NewG1()

	// obtain l generated from the pbc library
	// is a "small" number s.t. p + 1 = l*n
	l, err := parseLFromPBCParams(params)

	// choose random point P in G which becomes a generator for G of order N
	P = G1.Rand()
	P.PowBig(P, big.NewInt(0).Mul(l, big.NewInt(4)))
	// Make P a generate for the subgroup of order q1T

	// choose random Q in G1
	Q = G1.NewFieldElement()
	Q.PowBig(P, newCryptoRandom(N))
	Q.PowBig(Q, q2)

	// create public key with the generated groups
	pk := &PublicKey{pairing, G1, P, Q, N, T, polyBase, fpScaleBase, fpPrecision, deterministic}

	// create secret key
	sk := &SecretKey{q1, polyBase}

	if err != nil {
		panic("Couldn't generate key params!")
	}

	pk.computeEncodingTable()

	return pk, sk, err
}

// Encrypt a given plaintext (integer or rational) polynomial with the public key pk
func (pk *PublicKey) Encrypt(pt *Plaintext) *Ciphertext {

	encryptedCoefficients := make([]*pbc.Element, pt.Degree)

	for i := 0; i < pt.Degree; i++ {

		negative := pt.Coefficients[i] < 0
		if negative {
			positive := -1 * pt.Coefficients[i]
			coeff := big.NewInt(positive)
			encryptedCoefficients[i] = pk.ESubElements(pk.encryptZero(), pk.EncryptElement(coeff))
		} else {
			coeff := big.NewInt(pt.Coefficients[i])
			encryptedCoefficients[i] = pk.EncryptElement(coeff)
		}
	}

	return &Ciphertext{encryptedCoefficients, pt.Degree, pt.ScaleFactor, false}
}

// AInv returns the additive inverse of the level1 ciphertext
func (pk *PublicKey) AInv(ct *Ciphertext) *Ciphertext {

	if ct.L2 {
		return pk.aInvL2(ct)
	}

	degree := ct.Degree
	result := make([]*pbc.Element, degree)

	for i := degree - 1; i >= 0; i-- {
		result[i] = pk.ESubElements(pk.encryptZero(), ct.Coefficients[i])
	}

	return &Ciphertext{result, ct.Degree, ct.ScaleFactor, ct.L2}
}

// AInvElement returns the additive inverse of the level1 element
func (pk *PublicKey) AInvElement(el *pbc.Element) *pbc.Element {
	return pk.ESubElements(pk.encryptZero(), el)
}

// AInvElementL2 returns the additive inverse of the level2 element
func (pk *PublicKey) AInvElementL2(el *pbc.Element) *pbc.Element {
	return pk.ESubL2Elements(pk.ToDeterministicL2Element(pk.encryptZero()), el)
}

// EAdd adds two level 1 (non-multiplied) ciphertexts together and returns the result
func (pk *PublicKey) EAdd(ciphertext1 *Ciphertext, ciphertext2 *Ciphertext) *Ciphertext {

	if ciphertext1.L2 || ciphertext2.L2 {

		if !ciphertext1.L2 {
			return pk.EAddL2(pk.MakeL2(ciphertext1), ciphertext2)
		}

		if !ciphertext2.L2 {
			return pk.EAddL2(ciphertext1, pk.MakeL2(ciphertext2))
		}

		return pk.EAddL2(ciphertext1, ciphertext2)
	}

	ct1 := ciphertext1.Copy()
	ct2 := ciphertext2.Copy()
	ct1, ct2 = pk.alignCiphertexts(ct1, ct2, false)

	degree := int(math.Max(float64(ct1.Degree), float64(ct2.Degree)))
	result := make([]*pbc.Element, degree)

	for i := 0; i < degree; i++ {

		if ct2.Degree > i && ct1.Degree > i {
			result[i] = pk.EAddElements(ct1.Coefficients[i], ct2.Coefficients[i])
			continue
		}

		if i >= ct2.Degree {
			result[i] = ct1.Coefficients[i]
		}

		if i >= ct1.Degree {
			result[i] = ct2.Coefficients[i]
		}
	}

	return &Ciphertext{result, degree, ct1.ScaleFactor, ct1.L2}
}

// Decrypt the given ciphertext
func (sk *SecretKey) Decrypt(ct *Ciphertext, pk *PublicKey) *Plaintext {

	if ct.L2 {
		return sk.decryptL2(ct, pk)
	}

	size := ct.Degree
	plaintextCoeffs := make([]int64, size)

	for i := 0; i < ct.Degree; i++ {
		plaintextCoeffs[i] = sk.DecryptElement(ct.Coefficients[i], pk, false).Int64()
	}

	return &Plaintext{pk, plaintextCoeffs, size, ct.ScaleFactor}
}

func (pk *PublicKey) aInvL2(ct *Ciphertext) *Ciphertext {

	degree := ct.Degree
	result := make([]*pbc.Element, degree)

	for i := degree - 1; i >= 0; i-- {
		result[i] = pk.ESubL2Elements(pk.encryptZeroL2(), ct.Coefficients[i])
	}

	return &Ciphertext{result, ct.Degree, ct.ScaleFactor, ct.L2}
}

func (sk *SecretKey) DecryptElement(el *pbc.Element, pk *PublicKey, failed bool) *big.Int {

	gsk := pk.G1.NewFieldElement()
	csk := pk.G1.NewFieldElement()

	gsk.PowBig(pk.P, sk.Key)
	csk.PowBig(el, sk.Key)

	pt, err := pk.RecoverMessageWithDL(gsk, csk, false)

	if err != nil && !failed {
		elNeg := pk.AInvElement(el)
		return big.NewInt(0).Mul(big.NewInt(-1), sk.DecryptElement(elNeg, pk, true))
	}

	if err != nil && failed {
		panic("could not find discrete log!")
	}

	return pt
}

func (sk *SecretKey) DecryptElementL2(el *pbc.Element, pk *PublicKey, failed bool) *big.Int {

	gsk := pk.Pairing.NewGT().Pair(pk.P, pk.P)
	gsk.PowBig(gsk, sk.Key)

	csk := el.NewFieldElement()
	csk.PowBig(el, sk.Key)

	pt, err := pk.RecoverMessageWithDL(gsk, csk, true)

	if err != nil && !failed {
		elNeg := pk.AInvElementL2(el)
		return big.NewInt(0).Mul(big.NewInt(-1), sk.DecryptElementL2(elNeg, pk, true))
	}

	if err != nil {
		panic("could not find discrete log!")
	}

	return pt
}

// DecryptL2 a level 2 (multiplied) ciphertext C using secret key sk
func (sk *SecretKey) decryptL2(ct *Ciphertext, pk *PublicKey) *Plaintext {

	size := ct.Degree
	plaintextCoeffs := make([]int64, size)

	for i := 0; i < ct.Degree; i++ {
		plaintextCoeffs[i] = sk.DecryptElementL2(ct.Coefficients[i], pk, false).Int64()
	}

	return &Plaintext{pk, plaintextCoeffs, ct.Degree, ct.ScaleFactor}
}

// EAddL2 adds two level 2 (multiplied) ciphertexts together and returns the result
func (pk *PublicKey) EAddL2(ciphertext1 *Ciphertext, ciphertext2 *Ciphertext) *Ciphertext {

	ct1 := ciphertext1.Copy()
	ct2 := ciphertext2.Copy()
	ct1, ct2 = pk.alignCiphertexts(ct1, ct2, true)

	degree := int(math.Max(float64(ct1.Degree), float64(ct2.Degree)))
	result := make([]*pbc.Element, degree)

	for i := degree - 1; i >= 0; i-- {

		if i >= ct2.Degree {
			result[i] = ct1.Coefficients[i]
			continue
		}

		if i >= ct1.Degree {
			result[i] = ct2.Coefficients[i]
			continue
		}

		result[i] = pk.EAddL2Elements(ct1.Coefficients[i], ct2.Coefficients[i])
	}

	return &Ciphertext{result, degree, ct1.ScaleFactor, ct1.L2}
}

// EMultC multiplies a level 1 (non-multiplied) ciphertext with a plaintext constant
// and returns the result
func (pk *PublicKey) EMultC(ct *Ciphertext, constant *big.Float) *Ciphertext {

	if ct.L2 {
		return pk.eMultCL2(ct, constant)
	}

	return pk.eMultC(ct, constant)
}

func (pk *PublicKey) eMultC(ct *Ciphertext, constant *big.Float) *Ciphertext {

	isNegative := constant.Cmp(big.NewFloat(0.0)) < 0
	if isNegative {
		constant.Mul(constant, big.NewFloat(-1.0))
	}

	poly := pk.NewUnbalancedPlaintext(constant)

	degree := ct.Degree + poly.Degree
	result := make([]*pbc.Element, degree)

	zero := pk.G1.NewFieldElement()

	// set all coefficients to zero
	for i := 0; i < degree; i++ {
		result[i] = zero
	}

	for i := ct.Degree - 1; i >= 0; i-- {
		for k := poly.Degree - 1; k >= 0; k-- {
			index := i + k

			coeff := zero.NewFieldElement()
			coeff = pk.EMultCElement(ct.Coefficients[i], big.NewInt(poly.Coefficients[k]))
			result[index] = pk.EAddElements(result[index], coeff)
		}
	}

	product := &Ciphertext{result, degree, ct.ScaleFactor + poly.ScaleFactor, ct.L2}

	if isNegative {
		return pk.AInv(product)
	}

	return product
}

// EMultCL2 multiplies a level 2 (multiplied) ciphertext with a plaintext constant
// and returns the result
func (pk *PublicKey) eMultCL2(ct *Ciphertext, constant *big.Float) *Ciphertext {

	isNegative := constant.Cmp(big.NewFloat(0.0)) < 0
	if isNegative {
		constant.Mul(constant, big.NewFloat(-1.0))
	}

	poly := pk.NewUnbalancedPlaintext(constant)

	degree := ct.Degree + poly.Degree
	result := make([]*pbc.Element, degree)

	// set all coefficients to zero
	for i := 0; i < degree; i++ {
		result[i] = pk.Pairing.NewGT().NewFieldElement()
	}

	for i := ct.Degree - 1; i >= 0; i-- {
		for k := poly.Degree - 1; k >= 0; k-- {
			index := i + k

			coeff := pk.Pairing.NewGT().NewFieldElement()
			coeff = pk.EMultCElementL2(ct.Coefficients[i], big.NewInt(poly.Coefficients[k]))
			result[index] = pk.EAddL2Elements(result[index], coeff)
		}
	}

	product := &Ciphertext{result, degree, ct.ScaleFactor + poly.ScaleFactor, ct.L2}

	if isNegative {
		return pk.AInv(product)
	}

	return product
}

func (pk *PublicKey) EMultCElement(el *pbc.Element, constant *big.Int) *pbc.Element {

	res := el.NewFieldElement()
	res.PowBig(el, constant)

	if pk.Deterministic {
		r := newCryptoRandom(pk.N)
		q := el.NewFieldElement()
		q.MulBig(pk.Q, r)
		res.Mul(res, q)
	}

	return res
}

func (pk *PublicKey) EMultCElementL2(el *pbc.Element, constant *big.Int) *pbc.Element {

	res := pk.Pairing.NewGT().NewFieldElement()
	res.PowBig(el, constant)

	if !pk.Deterministic {
		r := newCryptoRandom(pk.N)
		pair := pk.Pairing.NewGT().NewFieldElement().Pair(pk.Q, pk.Q)
		pair.PowBig(pair, r)
		res.Mul(res, pair)
	}

	return res
}

// EMult multiplies two level 1 (non-multiplied) ciphertext together and returns the result
func (pk *PublicKey) EMult(ct1 *Ciphertext, ct2 *Ciphertext) *Ciphertext {

	degree := ct1.Degree + ct2.Degree
	result := make([]*pbc.Element, degree)

	// encrypt the padding zero coefficients
	for i := 0; i < degree; i++ {
		result[i] = pk.Pairing.NewGT().NewFieldElement()
	}

	for i := ct1.Degree - 1; i >= 0; i-- {
		for k := ct2.Degree - 1; k >= 0; k-- {
			index := i + k
			coeff := pk.EMultElements(ct1.Coefficients[i], ct2.Coefficients[k])
			result[index] = pk.EAddL2Elements(result[index], coeff)
		}
	}

	return &Ciphertext{result, degree, ct1.ScaleFactor + ct2.ScaleFactor, true}
}

func (pk *PublicKey) EMultElements(el1 *pbc.Element, el2 *pbc.Element) *pbc.Element {

	res := pk.Pairing.NewGT().NewFieldElement()
	res.Pair(el1, el2)

	if !pk.Deterministic {
		r := newCryptoRandom(pk.N)
		pair := pk.Pairing.NewGT().Pair(pk.Q, pk.Q)
		pair.PowBig(pair, r)
		res.Mul(res, pair)
	}

	return res
}

// MakeL2 moves a given ciphertext to the GT field
func (pk *PublicKey) MakeL2(ct *Ciphertext) *Ciphertext {

	one := pk.Encrypt(pk.NewPlaintext(big.NewFloat(1.0)))
	return pk.EMult(one, ct)
}

func (pk *PublicKey) toL2Element(el *pbc.Element) *pbc.Element {

	result := pk.Pairing.NewGT().NewFieldElement()
	result.Pair(el, pk.EncryptElement(big.NewInt(1)))

	r := newCryptoRandom(pk.N)
	pair := pk.Pairing.NewGT().Pair(pk.Q, pk.Q)
	pair.PowBig(pair, r)

	return result.Mul(result, pair)
}

func (pk *PublicKey) ToDeterministicL2Element(el *pbc.Element) *pbc.Element {

	result := pk.Pairing.NewGT().NewFieldElement()
	result.Pair(el, pk.EncryptDeterministic(big.NewInt(1)))
	return result
}

func (pk *PublicKey) EncryptDeterministic(x *big.Int) *pbc.Element {

	G := pk.G1.NewFieldElement()
	return G.PowBig(pk.P, x)
}

func (pk *PublicKey) EncryptElement(x *big.Int) *pbc.Element {

	G := pk.G1.NewFieldElement()
	G.PowBig(pk.P, x)

	r := newCryptoRandom(pk.N)
	H := pk.G1.NewFieldElement()
	H.PowBig(pk.Q, r)

	C := pk.G1.NewFieldElement()
	return C.Mul(G, H)
}

func (pk *PublicKey) RecoverMessageWithDL(gsk *pbc.Element, csk *pbc.Element, l2 bool) (*big.Int, error) {

	zero := gsk.NewFieldElement()

	if zero.Equals(csk) {
		return big.NewInt(0), nil
	}

	m, err := pk.getDL(csk, gsk, l2)

	if err != nil {
		return nil, err
	}
	return m, nil

}

func (pk *PublicKey) ESubElements(coeff1 *pbc.Element, coeff2 *pbc.Element) *pbc.Element {

	result := pk.G1.NewFieldElement()
	result.Div(coeff1, coeff2)
	if pk.Deterministic {
		return result // don't blind with randomness
	}

	rand := newCryptoRandom(pk.N)
	h1 := pk.G1.NewFieldElement()
	h1.PowBig(pk.Q, rand)
	return result.Mul(result, h1)
}

func (pk *PublicKey) ESubL2Elements(coeff1 *pbc.Element, coeff2 *pbc.Element) *pbc.Element {

	result := pk.Pairing.NewGT().NewFieldElement()
	result.Div(coeff1, coeff2)

	if pk.Deterministic {
		return result // don't hide with randomness
	}

	r := newCryptoRandom(pk.N)

	pair := pk.Pairing.NewGT().Pair(pk.Q, pk.Q)
	pair.PowBig(pair, r)
	return result.Mul(result, pair)
}

func (pk *PublicKey) EAddElements(coeff1 *pbc.Element, coeff2 *pbc.Element) *pbc.Element {

	result := pk.G1.NewFieldElement()
	result.Mul(coeff1, coeff2)

	if pk.Deterministic {
		return result // don't hide with randomness
	}

	rand := newCryptoRandom(pk.N)
	h1 := pk.G1.NewFieldElement()
	h1.PowBig(pk.Q, rand)

	return result.Mul(result, h1)
}

func (pk *PublicKey) EAddL2Elements(coeff1 *pbc.Element, coeff2 *pbc.Element) *pbc.Element {

	result := pk.Pairing.NewGT().NewFieldElement()
	result.Mul(coeff1, coeff2)

	if pk.Deterministic {
		return result // don't hide with randomness
	}

	r := newCryptoRandom(pk.N)
	pair := pk.Pairing.NewGT().Pair(pk.Q, pk.Q)
	pair.PowBig(pair, r)

	return result.Mul(result, pair)
}

func (pk *PublicKey) EPolyEval(ct *Ciphertext) *pbc.Element {
	acc := pk.EncryptDeterministic(big.NewInt(0))
	x := big.NewInt(int64(pk.PolyBase))

	for i := ct.Degree - 1; i >= 0; i-- {
		acc = pk.EMultCElement(acc, x)
		acc = pk.EAddElements(acc, ct.Coefficients[i])
	}

	return acc
}

func (pk *PublicKey) alignCiphertexts(ct1 *Ciphertext, ct2 *Ciphertext, level2 bool) (*Ciphertext, *Ciphertext) {

	if ct1.ScaleFactor > ct2.ScaleFactor {
		diff := ct1.ScaleFactor - ct2.ScaleFactor

		ct2 = pk.EMultC(ct2, big.NewFloat(math.Pow(float64(pk.FPScaleBase), float64(diff))))
		ct2.ScaleFactor = ct1.ScaleFactor

	} else if ct2.ScaleFactor > ct1.ScaleFactor {
		// flip the ciphertexts
		return pk.alignCiphertexts(ct2, ct1, level2)
	}

	return ct1, ct2
}

func (pk *PublicKey) encryptZero() *pbc.Element {
	return pk.EncryptElement(big.NewInt(0))
}

func (pk *PublicKey) encryptZeroL2() *pbc.Element {

	zero := pk.encryptZero()

	result := pk.Pairing.NewGT().NewFieldElement()
	result.Pair(zero, zero)

	r := newCryptoRandom(pk.N)
	pair := pk.Pairing.NewGT().Pair(pk.Q, pk.Q)
	pair.PowBig(pair, r)

	result.Mul(result, pair)

	return result
}

// generates a new random number < max
func newCryptoRandom(max *big.Int) *big.Int {
	rand, err := rand.Int(rand.Reader, max)
	if err != nil {
		log.Println(err)
	}

	return rand
}

// TOTAL HACK to access the generated "l" in the C struct
// which the PBC library holds. The golang wrapper has
// no means of accessing the struct variable without
// knowing the exact memory mapping. Better approach
// would be to either compute l on the fly or figure
// out the memory mapping between the C struct and
// golang equivalent
func parseLFromPBCParams(params *pbc.Params) (*big.Int, error) {

	paramsStr := params.String()
	lStr := paramsStr[strings.Index(paramsStr, "l")+2 : len(paramsStr)-1]
	lInt, err := strconv.ParseInt(lStr, 10, 64)
	if err != nil {
		return nil, err
	}

	return big.NewInt(lInt), nil
}

func (c *Ciphertext) String() string {

	str := ""
	for _, coeff := range c.Coefficients {
		str += coeff.String() + "\n"
	}

	return str
}
