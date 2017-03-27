package contract

import (
	. "DNA/common"
	"DNA/vm"
	"io"
	"bytes"
	"DNA/common/serialization"
	. "DNA/errors"
	"errors"
)

//Contract address is the hash of contract program .
//which be used to control asset or indicate the smart contract address �?


//Contract include the program codes with parameters which can be executed on specific evnrioment
type Contract struct {

	//the contract program code,which will be run on VM or specific envrionment
	Code []byte

	//the Contract Parameter type list
	// describe the number of contract program parameters and the parameter type
	Parameters []ContractParameterType

	//The program hash as contract address
	ProgramHash Uint160

	//owner's pubkey hash indicate the owner of contract
	OwnerPubkeyHash Uint160

}

func (c *Contract) IsStandard() bool {
	if len(c.Code) != 35 {
		return false
	}
	if c.Code[0] != 33 || c.Code[34] != byte(vm.OP_CHECKSIG) {
		return false
	}
	return true
}

func (c *Contract) IsMultiSigContract() bool {
	var m int16 = 0
	var n int16 = 0
	i := 0

	if len(c.Code) < 37 {return false}
	if c.Code[i] > byte(vm.OP_16) {return false}
	if c.Code[i] < byte(vm.OP_1) && c.Code[i] != 1 && c.Code[i] != 2 {
		return false
	}

	switch c.Code[i] {
	case 1:
		i++
		m = int16(c.Code[i])
		i++
		break
	case 2:
		i++
		m = BytesToInt16(c.Code[i:])
		i += 2
		break
	default:
		m = int16(c.Code[i]) - 80
		i++
		break
	}

	if m < 1 || m > 1024 {return false}

	for c.Code[i] == 33 {
		i += 34
		if len(c.Code) <= i {return false}
		n++
	}
	if n < m || n > 1024 {return false}

	switch c.Code[i] {
	case 1:
		i++
		if n != int16(c.Code[i]) {return false}
		i++
		break
	case 2:
		i++
		if n != BytesToInt16(c.Code[i:]) {return false}
		i += 2
		break
	default:
		if n != (int16(c.Code[i]) - 80) {return false}
		i++
		break
	}

	if c.Code[i] != byte(vm.OP_CHECKMULTISIG) {return false}
	i++
	if len(c.Code) != i {return false}

	return true
}

func (c *Contract) GetType() ContractType{
	if c.IsStandard() {
		return SignatureContract
	}
	if c.IsMultiSigContract() {
		return MultiSigContract
	}
	return CustomContract
}

func (c *Contract) Deserialize(r io.Reader) error {
	c.OwnerPubkeyHash.Deserialize(r)

	p,err := serialization.ReadVarBytes(r)
	if err != nil {
		return err
	}
	c.Parameters = ByteToContractParameterType(p)

	c.Code,err = serialization.ReadVarBytes(r)
	if err != nil {
		return err
	}

	return nil
}

func (c *Contract) Serialize(w io.Writer) error {
	len,err := c.OwnerPubkeyHash.Serialize(w)
	if err != nil {
		return err
	}
	if len != 20 {
		return NewDetailErr(errors.New("PubkeyHash.Serialize(): len != len(Uint160)"), ErrNoCode, "")
	}

	err = serialization.WriteVarBytes(w,ContractParameterTypeToByte(c.Parameters))
	if err != nil {
		return err
	}

	err = serialization.WriteVarBytes(w,c.Code)
	if err != nil {
		return err
	}

	return nil
}

func (c *Contract) ToArray() []byte {
	w := new(bytes.Buffer)
	c.Serialize(w)

	return w.Bytes()
}

func ContractParameterTypeToByte( c [] ContractParameterType ) []byte {
	b := make( []byte, len(c) )

	for i:=0; i<len(c); i++ {
		b[i] = byte(c[i])
	}

	return b
}

func ByteToContractParameterType( b []byte ) []ContractParameterType {
	c := make( []ContractParameterType, len(b) )

	for i:=0; i<len(b); i++ {
		c[i] = ContractParameterType(b[i])
	}

	return c
}


