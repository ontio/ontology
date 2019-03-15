package snark

import (
	"encoding/hex"
	"math/big"
	"os"
	"testing"

	"github.com/kunxian-xia/bn256"
	"github.com/ontio/ontology/common"
)

func hexstringToBigInt(h string) *big.Int {
	bytes, _ := hex.DecodeString(h)
	return new(big.Int).SetBytes(bytes)
}

func conv_bytes_into_bit_vectors(bs []byte) []byte {
	bits := make([]byte, len(bs)*8)
	for i := 0; i < len(bs); i++ {
		c := uint(bs[i])
		for j := 0; j < 8; j++ {
			bits[i*8+j] = byte((c >> uint((7 - j))) & 1)
		}
	}
	return bits
}

func conv_uint64_into_LE_bytes(val uint64) []byte {
	bs := make([]byte, 8)
	for i := uint(0); i < 8; i++ {
		bs[i] = byte((val >> (8 * i)) & 0xff)
	}
	return bs
}

func pack_bytes_into_field_elements(bs []byte, chunk_size int) []*big.Int {
	numBits := len(bs) * 8
	numChunk := (numBits + chunk_size - 1) / chunk_size

	elements := make([]*big.Int, numChunk)
	for i := 0; i < numChunk; i++ {
		elements[i] = new(big.Int)
		for j := 0; j < chunk_size; j++ {
			pos := i*chunk_size + j
			if pos >= numBits {
				break
			}
			bit := uint(bs[pos/8])
			bit = bit >> (7 - uint(pos%8))
			bit = bit & 0x01
			elements[i].SetBit(elements[i], j, bit)
		}
	}
	return elements
}

// func TestPack(t *testing.T) {
// 	bs := []byte{1, 2, 3, 4, 5}
// 	fs := pack_bytes_into_field_elements(bs, 254)
// 	for i := 0; i < len(fs); i++ {
// 		t.Logf(fs[i].String())
// 	}
// 	t.Fatal("stop")
// }
func TestPHGR13Verify(t *testing.T) {
	// we are testing against a real world case
	// i.e. zk-SNARKs proof contained in a ZCash mainnet transaction
	// whose hash is ec31a1b3e18533702c74a67d91c49d622717bd53d6192c5cb23b9bdf080416a5.
	// Moreover, this transaction is included in height 396.
	// Ref: https://api.zcha.in/v2/mainnet/transactions/ec31a1b3e18533702c74a67d91c49d622717bd53d6192c5cb23b9bdf080416a5

	// vk
	vk_file, err := os.Open("sprout-verifying.key")
	if err != nil {
		t.Fatal(err)
	}
	vk, err := parseVK(vk_file)
	if err != nil {
		t.Fatalf("parse vk failed, err: %x", err)
	}

	// proof
	proof_file, err := os.Open("proof.txt")
	if err != nil {
		t.Fatal(err)
	}
	proof_raw := make([]byte, 2*(33*7+65))
	_, err = proof_file.Read(proof_raw)
	if err != nil {
		t.Fatal(err)
	}
	proof_raw, _ = hex.DecodeString(string(proof_raw))

	proof, err := importProof(proof_raw)
	if err != nil {
		t.Fatal(err)
	}

	xx, _ := new(big.Int).SetString("20507014976900324884923703462229212939510025188133599277134408844142237392307", 10)
	xy, _ := new(big.Int).SetString("539045453165532223624174985214626049922380177513464865201634923239231964598", 10)
	yx, _ := new(big.Int).SetString("17706169778760270831199133527036782630364426470105034702709971627848861923912", 10)
	yy, _ := new(big.Int).SetString("12178221303165765388438472058234866951705348968791707113588519754197989138595", 10)

	b := make([]byte, 0, 128)
	b = append(b, xx.Bytes()...)
	b = append(b, xy.Bytes()...)
	b = append(b, yx.Bytes()...)
	b = append(b, yy.Bytes()...)

	source := common.NewZeroCopySource(b)

	proof.b = new(bn256.G2)
	err = deserializeG2(proof.b, source)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("b: %s\n", proof.b.String())

	// public input
	_input := make([]*big.Int, 9)
	_inputStrs := []string{
		"11893887518801564238850113243068155191401763535822078310914655246254174921707",
		"9039742628274832857146315176202079824763880684544058044764009859702372701908",
		"7864248849999267529324215987921491632294157863019983191999113732927809771441",
		"2886983623257678406932083534975273655277211437585781522465101031866117927530",
		"1639613592978633992206850322587892881255594351774222883941421746126476816445",
		"5902043119256669211364401966461491601894820710756687540191805850512824202436",
		"13692185839566206949758987046107079401517252355870659294323573892338548513162",
		"213567272714802366240312308317683913515756890632602759628885800370159516315",
		"170484577853289",
	}
	for i := 0; i < len(_input); i++ {
		_input[i], _ = new(big.Int).SetString(_inputStrs[i], 10)
	}

	valid, err := verify(vk, proof, _input)
	if err != nil {
		t.Fatal(err)
	}
	if !valid {
		t.Fatal("phgr verify failed")
	}
}
