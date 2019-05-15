package bgn

import (
	"github.com/Nik-U/pbc"
)

type Ciphertext struct {
	Coefficients []*pbc.Element // coefficients in the plaintext or ciphertext poly
	Degree       int
	ScaleFactor  int
	L2           bool // whether ciphertext is level2
}

// NewCiphertext generates a  new ciphertext...duh
func NewCiphertext(coefficients []*pbc.Element, degree int, scaleFactor int, l2 bool) *Ciphertext {

	return &Ciphertext{coefficients, degree, scaleFactor, l2}
}
