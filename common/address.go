package common

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"github.com/Ontology/common/log"
	. "github.com/Ontology/errors"
	"github.com/itchyny/base58-go"
	"io"
	"math/big"
)

const AddrLen int = 20

type Address [AddrLen]uint8

func (u *Address) ToArray() []byte {
	var x []byte = make([]byte, AddrLen)
	for i := 0; i < 20; i++ {
		x[i] = byte(u[i])
	}

	return x
}

func (u *Address) ToString() string {
	return hex.EncodeToString(u.ToArray())
}

func (u *Address) Serialize(w io.Writer) (int, error) {
	b_buf := bytes.NewBuffer([]byte{})
	binary.Write(b_buf, binary.LittleEndian, u)

	len, err := w.Write(b_buf.Bytes())

	if err != nil {
		return 0, err
	}

	return len, nil
}

func (f *Address) Deserialize(r io.Reader) error {
	p := make([]byte, AddrLen)
	n, err := r.Read(p)

	if n <= 0 || err != nil {
		return err
	}

	b_buf := bytes.NewBuffer(p)
	binary.Read(b_buf, binary.LittleEndian, f)

	return nil
}

func (f *Address) ToBase58() string {
	data := append([]byte{0x41}, f[:]...)
	temp := sha256.Sum256(data)
	temps := sha256.Sum256(temp[:])
	data = append(data, temps[0:4]...)

	bi := new(big.Int).SetBytes(data).String()
	encoded, _ := base58.BitcoinEncoding.Encode([]byte(bi))
	return string(encoded)
}

func Uint160ParseFromBytes(f []byte) (Address, error) {
	if len(f) != AddrLen {
		return Address{}, NewDetailErr(errors.New("[Common]: Uint160ParseFromBytes err, len != 20"), ErrNoCode, "")
	}

	var hash [20]uint8
	for i := 0; i < 20; i++ {
		hash[i] = f[i]
	}
	return Address(hash), nil
}

func AddressFromBase58(encoded string) (Address, error) {
	decoded, err := base58.BitcoinEncoding.Decode([]byte(encoded))
	if err != nil {
		return Address{}, err
	}

	x, _ := new(big.Int).SetString(string(decoded), 10)
	log.Tracef("[ToAddress] x: ", x.Bytes())

	ph, err := Uint160ParseFromBytes(x.Bytes()[1:21])
	if err != nil {
		return Address{}, err
	}

	log.Tracef("[AddressToProgramHash] programhash: %x", ph[:])

	addr := ph.ToBase58()

	log.Tracef("[AddressToProgramHash] encoded: %s", addr)

	if addr != encoded {
		return Address{}, errors.New("[AddressFromBase58]: decode encoded verify failed.")
	}

	return ph, nil
}
