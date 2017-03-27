package crypto

import (
	. "DNA/common"
	"crypto/sha256"
	"fmt"
	"testing"
)

func TestHash(t *testing.T) {

	var data []Uint256
	a1 := Uint256(sha256.Sum256([]byte("a")))
	a2 := Uint256(sha256.Sum256([]byte("b")))
	a3 := Uint256(sha256.Sum256([]byte("c")))
	a4 := Uint256(sha256.Sum256([]byte("d")))
	a5 := Uint256(sha256.Sum256([]byte("e")))
	data = append(data, a1)
	data = append(data, a2)
	data = append(data, a3)
	data = append(data, a4)
	data = append(data, a5)
	x, _ := ComputeRoot(data)
	fmt.Printf("[Root Hash]:%x\n", x)

}
