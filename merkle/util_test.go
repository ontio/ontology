package merkle

import "testing"

func TestBitOp(t *testing.T) {
	base := uint32(100001)
	for i := 0; i < 1000; i++ {
		n := base + uint32(i)
		if countBit(n) != countBitOld(n) {
			t.Fatal("countBit check fail", n)
		}
		if highBit(n) != highBitOld(n) {
			t.Fatal("highBit check fail", n)
		}
	}
}

func countBitOld(num uint32) uint {
	var count uint
	for num != 0 {
		num &= num - 1
		count += 1
	}
	return count
}

func highBitOld(num uint32) uint {
	var hiBit uint
	for num != 0 {
		num >>= 1
		hiBit += 1
	}
	return hiBit
}
