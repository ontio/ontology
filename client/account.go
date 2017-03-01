package client

import (
	"GoOnchain/crypto"
	. "GoOnchain/common"
	. "GoOnchain/errors"
	"errors"
)

type Account struct {
	PrivateKey []byte
	PublicKey *crypto.PubKey
	PublicKeyHash Uint160
}

func NewAccount() (*Account, error){

	priKey,pubKey,_ := crypto.GenKeyPair()
	temp,err := pubKey.EncodePoint(true)
	if err !=nil{
		return nil,NewDetailErr(err, ErrNoCode, "[Contract],CreateSignatureContract failed.")
	}
	hash ,err  := ToCodeHash(temp)
	if err !=nil{
		return nil,NewDetailErr(err, ErrNoCode, "[Contract],CreateSignatureContract failed.")
	}
	return &Account{
		PrivateKey: priKey,
		PublicKey: &pubKey,
		PublicKeyHash: hash,
	},nil
}

func NewAccountWithPrivatekey(privateKey []byte) (*Account, error){
	privKeyLen := len(privateKey)

	if privKeyLen != 32 && privKeyLen != 96 && privKeyLen != 104 {
		return nil,errors.New("Invalid private Key.")
	}

	// set public key
	pubKey := crypto.NewPubKey(privateKey)
	//priKey,pubKey,_ := crypto.GenKeyPair()
	temp,err := pubKey.EncodePoint(true)
	if err !=nil{
		return nil,NewDetailErr(err, ErrNoCode, "[Contract],CreateSignatureContract failed.")
	}
	hash ,err  := ToCodeHash(temp)
	if err !=nil{
		return nil,NewDetailErr(err, ErrNoCode, "[Contract],CreateSignatureContract failed.")
	}
	return &Account{
		PrivateKey: privateKey,
		PublicKey: pubKey,
		PublicKeyHash: hash,
	},nil
}

//get signer's private key
func (ac *Account) PrivKey() []byte{
	return ac.PrivateKey
}

//get signer's public key
func (ac *Account) PubKey() *crypto.PubKey {
	return ac.PublicKey
}
/*
func (ac *Account) ToAddress(scriptHash Uint160) string {
	//fmt.Printf( "%x\n", scriptHash )
	data := append( []byte{23}, scriptHash.ToArray()... )
	temp := sha256.Sum256(data)
	temps:= sha256.Sum256(temp[:])
	data = append( data, temps[0:4]... )

	data = append( []byte{0x00}, data... )
	//fmt.Printf( "%x\n", data )

	// Reverse
	// base58.EncodeBig alread Reverse

	//var dataint = make([]int,len(data))
	//for i:=0; i<len(data); i++ {
	//	dataint[i] = int(data[i])
	//}
	//for i:=0; i<len(dataint); i++ {
	//	data[i] = byte(dataint[len(dataint)-1-i])
	//}
	//fmt.Printf( "%x\n", data )

	bi := new( big.Int )
	bi.SetBytes( data )
	var dst []byte
	dst = base58.EncodeBig( dst, bi )
	//fmt.Printf( "%x\n", dst )

	return string(dst[:])
}
*/