package serialization

import (
	"io"
)

//SerializableData describe the data need be serialized.
type SerializableData interface {

	//Write data to writer
	Serialize(w io.Writer)

	//read data to reader
	Deserialize(r io.Reader)


}

func WriteDataList(w io.Writer, list []SerializableData)  error {
	len := uint64(len(list))
	WriteVarInt(w,len)

	for _, data := range list {
		data.Serialize(w)
	}

	return nil
}

func WriteVarInt(w io.Writer, intval interface{}) (uint64, error){
	//TODO: implement WriteVarInt

	return 0,nil
}

func WriteVarString(w io.Writer, val *string) (int, error){
	//TODO: implement WriteVarString

	return 0,nil
}

func WriteVarBytes(w io.Writer, val []byte) (int, error){
	len := uint64(len(val))
	WriteVarInt(w,len)
	return w.Write(val)
}

func ReadVarInt(w io.Reader) (int){
	//TODO: implement ReadVarInt

	return 0
}

func ReadUint(r io.Reader) (uint) {
	//TODO: implement ReadUint

	return 0
}

func ReadUint64(r io.Reader) (uint64) {
	//TODO: implement ReadUint32

	return 0
}

