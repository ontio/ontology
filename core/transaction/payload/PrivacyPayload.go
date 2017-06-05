package payload

import (
	"DNA/common/serialization"
	"DNA/crypto"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"errors"
	"io"
	mrand "math/rand"
	"time"
)

type EncryptedPayloadType byte

type EncryptedPayload []byte

type PayloadEncryptType byte

type PayloadEncryptAttr interface {
	Serialize(w io.Writer) error
	Deserialize(r io.Reader) error
	Encrypt(msg []byte, keys interface{}) ([]byte, error)
	Decrypt(msg []byte, keys interface{}) ([]byte, error)
}

const (
	ECDH_AES256 PayloadEncryptType = 0x01
)
const (
	RawPayload EncryptedPayloadType = 0x01
)

type PrivacyPayload struct {
	PayloadType EncryptedPayloadType
	Payload     EncryptedPayload
	EncryptType PayloadEncryptType
	EncryptAttr PayloadEncryptAttr
}

func (pp *PrivacyPayload) Data() []byte {
	//TODO: implement PrivacyPayload.Data()
	return []byte{0}
}

func (pp *PrivacyPayload) Serialize(w io.Writer) error {
	w.Write([]byte{byte(pp.PayloadType)})
	err := serialization.WriteVarBytes(w, pp.Payload)
	if err != nil {
		return err
	}
	w.Write([]byte{byte(pp.EncryptType)})
	err = pp.EncryptAttr.Serialize(w)

	return err
}

func (pp *PrivacyPayload) Deserialize(r io.Reader) error {
	var PayloadType [1]byte
	_, err := io.ReadFull(r, PayloadType[:])
	if err != nil {
		return err
	}
	pp.PayloadType = EncryptedPayloadType(PayloadType[0])

	Payload, err := serialization.ReadVarBytes(r)
	if err != nil {
		return err
	}
	pp.Payload = Payload

	var encryptType [1]byte
	_, err = io.ReadFull(r, encryptType[:])
	if err != nil {
		return err
	}
	pp.EncryptType = PayloadEncryptType(encryptType[0])

	switch pp.EncryptType {
	case ECDH_AES256:
		pp.EncryptAttr = new(EcdhAes256)
	default:
		return errors.New("unknown EncryptType")
	}
	err = pp.EncryptAttr.Deserialize(r)

	return err
}

type EcdhAes256 struct {
	FromPubkey *crypto.PubKey
	ToPubkey   *crypto.PubKey
	Nonce      []byte
}

func (ea *EcdhAes256) Serialize(w io.Writer) error {
	err := ea.FromPubkey.Serialize(w)
	if err != nil {
		return err
	}
	err = ea.ToPubkey.Serialize(w)
	if err != nil {
		return err
	}
	err = serialization.WriteVarBytes(w, ea.Nonce)
	return err
}
func (ea *EcdhAes256) Deserialize(r io.Reader) error {
	ea.FromPubkey = new(crypto.PubKey)
	err := ea.FromPubkey.DeSerialize(r)
	if err != nil {
		return err
	}

	ea.ToPubkey = new(crypto.PubKey)
	err = ea.ToPubkey.DeSerialize(r)
	if err != nil {
		return err
	}

	nonce, err := serialization.ReadVarBytes(r)
	if err != nil {
		return err
	}
	ea.Nonce = nonce
	return nil
}

func (ea *EcdhAes256) Encrypt(msg []byte, keys interface{}) ([]byte, error) {
	var key []byte
	switch keys.(type) {
	case []byte:
		key = keys.([]byte)
	default:
		return []byte{}, errors.New("The keys error")
	}

	priv, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return []byte{}, err
	}
	x, _ := priv.Curve.ScalarMult(ea.ToPubkey.X, ea.ToPubkey.Y, key)
	aesKey := make([]byte, 32)
	copy(aesKey[32-len(x.Bytes()):], x.Bytes())

	iv := make([]byte, 16)
	r := mrand.New(mrand.NewSource(time.Now().UnixNano()))
	for i := 0; i < 16; i++ {
		iv[i] = byte(r.Intn(256))
	}

	paddingData := crypto.PKCS5Padding(msg, 16)
	encryption, err := crypto.AesEncrypt(paddingData, aesKey, iv)
	if err != nil {
		return []byte{}, err
	}
	ea.Nonce = iv

	return encryption, nil
}

func (ea *EcdhAes256) Decrypt(msg []byte, keys interface{}) ([]byte, error) {
	var key []byte
	switch keys.(type) {
	case []byte:
		key = keys.([]byte)
	default:
		return []byte{}, errors.New("The keys error")
	}

	priv, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return []byte{}, err
	}
	x, _ := priv.Curve.ScalarMult(ea.FromPubkey.X, ea.FromPubkey.Y, key)
	aesKey := make([]byte, 32)
	copy(aesKey[32-len(x.Bytes()):], x.Bytes())

	decryption, _ := crypto.AesDecrypt(msg, aesKey, ea.Nonce)
	result := crypto.PKCS5UnPadding(decryption)

	return result, nil
}
