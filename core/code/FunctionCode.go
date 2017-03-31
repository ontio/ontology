package code

import (
	"DNA/common/log"
	. "DNA/common"
	. "DNA/core/contract"
	"DNA/common/serialization"
	"fmt"
	"io"
)

type FunctionCode struct {
	// Contract Code
	Code []byte

	// Contract parameter type list
	ParameterTypes []ContractParameterType

	// Contract return type list
	ReturnTypes []ContractParameterType
}

// method of SerializableData
func (fc *FunctionCode) Serialize(w io.Writer) error {
	err := serialization.WriteVarBytes(w,ContractParameterTypeToByte(fc.ParameterTypes))
	if err != nil {
		return err
	}

	err = serialization.WriteVarBytes(w,fc.Code)
	if err != nil {
		return err
	}

	return nil
}

// method of SerializableData
func (fc *FunctionCode) Deserialize(r io.Reader) error {
	p,err := serialization.ReadVarBytes(r)
	if err != nil {
		return err
	}
	fc.ParameterTypes = ByteToContractParameterType(p)

	fc.Code,err = serialization.ReadVarBytes(r)
	if err != nil {
		return err
	}

	return nil
}

// method of ICode
// Get code
func (fc *FunctionCode) GetCode() []byte {
	return fc.Code
}

// method of ICode
// Get the list of parameter value
func (fc *FunctionCode) GetParameterTypes() []ContractParameterType {
	return fc.ParameterTypes
}

// method of ICode
// Get the list of return value
func (fc *FunctionCode) GetReturnTypes() []ContractParameterType {
	return fc.ReturnTypes
}

// method of ICode
// Get the hash of the smart contract
func (fc *FunctionCode) CodeHash() Uint160 {
	hash,err := ToCodeHash(fc.Code)
	if err != nil {
		log.Debug( fmt.Sprintf("[FunctionCode] ToCodeHash err=%s",err) )
		return Uint160{0}
	}

	return hash
}
