package common

import (
	"strings"
)

const (
	ErrDBNotFound    = "leveldb: not found"
)

func IsLevelDBNotFound(err error) bool{
	if err == nil {
		return false
	}
	return strings.EqualFold(err.Error(), ErrDBNotFound)
}
