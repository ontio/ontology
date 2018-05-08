package common

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestCompactUint(t *testing.T) {
	a := uint64(1200000)
	compactValue := SetCompactUint(a)
	unCompactValue, _ := GetCompactUint(compactValue)
	assert.Equal(t, unCompactValue, a)
}
