package states

import (
	"io"
	. "github.com/Ontology/common/serialization"
)

type StateBase struct {
	StateVersion byte
}

func(this *StateBase) Serialize(w io.Writer) error {
	WriteByte(w, this.StateVersion)
	return nil
}

func(this *StateBase) Deserialize(r io.Reader) error {
	if this == nil {
		this = new(StateBase)
	}
	stateVersion, err := ReadByte(r)
	if err != nil {
		return err
	}
	this.StateVersion = stateVersion
	return nil
}

