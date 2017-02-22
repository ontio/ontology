package crypto

import (
	. "GoOnchain/errors"
	"crypto/ecdsa"
	"crypto/rand"
	"errors"
	"io"
	"math/big"
)

const (
	INFINITYLEN      = 1
	FLAGLEN          = 1
	XORYVALUELEN     = 32
	COMPRESSEDLEN    = 33
	NOCOMPRESSEDLEN  = 65
	COMPEVENFLAG     = 0x02
	COMPODDFLAG      = 0x03
	NOCOMPRESSEDFLAG = 0x04
)

type PubKey ECPoint

func (e *PubKey) Serialize(w io.Writer) {
	//TODO: implement PubKey.serialize
}

func (e *PubKey) DeSerialize(r io.Reader) error {
	//TODO
	return nil
}

func isEven(k *big.Int) bool {
	z := big.NewInt(0)
	z.Mod(k, big.NewInt(2))
	if z.Int64() == 0 {
		return true
	}
	return false
}

// EncodePoint is used for compressing PublicKey for less space used as same as which in bitcoin.
func (e *PubKey) EncodePoint(isCommpressed bool) ([]byte, error) {
	//if X is infinity, then Y cann't be computed, so here used "||"
	if nil == e.X || nil == e.Y {
		return nil, NewDetailErr(errors.New("The PubKey is an infinity point"), ErrNoCode, "")
	}

	var encodedData []byte

	if isCommpressed {
		encodedData = make([]byte, COMPRESSEDLEN)
	} else {
		encodedData = make([]byte, NOCOMPRESSEDLEN)

		yBytes := e.Y.Bytes()
		copy(encodedData[NOCOMPRESSEDLEN-len(yBytes):], yBytes)
	}

	xBytes := e.X.Bytes()
	copy(encodedData[FLAGLEN:COMPRESSEDLEN], xBytes)

	if isCommpressed {
		if isEven(e.Y) {
			encodedData[0] = COMPEVENFLAG
		} else {
			encodedData[0] = COMPODDFLAG
		}
	} else {
		encodedData[0] = NOCOMPRESSEDFLAG
	}

	return encodedData, nil
}

func NewPubKey(prikey []byte) *PubKey {
	//TODO: NewPubKey
	return nil
}

func GenPrivKey() []byte {
	return nil
}

//FIXME, does the privkey need base58 encoding?
//This generates a public & private key pair
func GenKeyPair() ([]byte, PubKey, error) {
	pubkey := new(PubKey)
	privatekey := new(ecdsa.PrivateKey)
	privatekey, err := ecdsa.GenerateKey(Crypto.curve, rand.Reader)
	if err != nil {
		return nil, *pubkey, errors.New("Generate key pair error")
	}

	privkey := privatekey.D.Bytes()
	pubkey.X = privatekey.PublicKey.X
	pubkey.Y = privatekey.PublicKey.Y
	return privkey, *pubkey, nil
}

func DecodePoint(encoded []byte) *PubKey {
	//TODO: DecodePoint
	return nil
}

type PubKeySlice []*PubKey

func (p PubKeySlice) Len() int { return len(p) }
func (p PubKeySlice) Less(i, j int) bool {
	//TODO:PubKeySlice Less
	return false
}
func (p PubKeySlice) Swap(i, j int) {
	//TODO:PubKeySlice Swap
}
