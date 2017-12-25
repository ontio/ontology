package states

import (
	"io"
)

type IStateValue interface {
	Serialize(w io.Writer) error
	Deserialize(r io.Reader) error
}