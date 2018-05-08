package states

import (
	"bytes"
	"testing"
)

func TestStorageItem_Serialize_Deserialize(t *testing.T) {

	item := &StorageItem{
		StateBase: StateBase{StateVersion: 1},
		Value:     []byte{1},
	}

	bf := new(bytes.Buffer)
	if err := item.Serialize(bf); err != nil {
		t.Fatalf("StorageItem serialize error: %v", err)
	}

	var storage = new(StorageItem)
	if err := storage.Deserialize(bf); err != nil {
		t.Fatalf("StorageItem deserialize error: %v", err)
	}
}
