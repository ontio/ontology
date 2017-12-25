package neovm

import "testing"

func TestCat(t *testing.T) {
	a := []byte{1, 2, 3, 4, 5, 6, 7, 8}
	b := a[0:3]
	c := []byte{7, 8}
	d := append(b, c...)
	t.Log("d", d)
}
