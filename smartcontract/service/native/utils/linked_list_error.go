package utils

import (
	"encoding/hex"
	"fmt"
)

type linkedListError struct {
	m string          // error message
	k []byte          // key of the list node
	n *LinkedlistNode // list node
	o string          // operation
}

func (this *linkedListError) Error() string {
	return fmt.Sprintf("linked list %s error: %s (key: %s)",
		this.o, this.m, hex.EncodeToString(this.k))
}

func newInsertError(msg string, key []byte, node *LinkedlistNode) error {
	return &linkedListError{m: msg, k: key, n: node, o: "insert"}
}

func newDeleteError(msg string, key []byte) error {
	return &linkedListError{m: msg, k: key, n: nil, o: "delete"}
}
