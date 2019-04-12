package common

type Serializable interface {
	Serialization(sink *ZeroCopySink)
}

func SerializeToBytes(values ...Serializable) []byte {
	sink := NewZeroCopySink(0)
	for _, val := range values {
		val.Serialization(sink)
	}

	return sink.Bytes()
}
