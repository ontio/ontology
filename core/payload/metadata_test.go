package payload

import (
	"bytes"
	"testing"

	"github.com/magiconair/properties/assert"
)

func TestMetaDataCode(t *testing.T) {
	meta := NewDefaultMetaData()
	bf := new(bytes.Buffer)
	err := meta.Serialize(bf)
	if err != nil {
		t.Fatal(err)
	}
	newMeta := &MetaDataCode{}
	err = newMeta.Deserialize(bf)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, meta, newMeta)
}
