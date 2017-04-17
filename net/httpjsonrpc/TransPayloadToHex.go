package httpjsonrpc

import (
	. "DNA/common"
	"DNA/core/asset"
	. "DNA/core/contract"
	. "DNA/core/transaction"
	"DNA/core/transaction/payload"
	"DNA/crypto"
)

type PayloadInfo interface {
	Data() []byte
}

//implement PayloadInfo define BookKeepingInfo
type BookKeepingInfo struct {
}

func (dc *BookKeepingInfo) Data() []byte {
	return []byte{0}
}

//implement PayloadInfo define DeployCodeInfo
type FunctionCodeInfo struct {
	Code           string
	ParameterTypes string
	ReturnTypes    string
}

type DeployCodeInfo struct {
	Code        *FunctionCodeInfo
	Name        string
	CodeVersion string
	Author      string
	Email       string
	Description string
}

func (dc *DeployCodeInfo) Data() []byte {
	return []byte{0}
}

//implement PayloadInfo define IssueAssetInfo
type IssueAssetInfo struct {
}

func (a *IssueAssetInfo) Data() []byte {
	return []byte{0}
}

//implement PayloadInfo define RegisterAssetInfo
type RegisterAssetInfo struct {
	Asset      *asset.Asset
	Amount     Fixed64
	Issuer     *crypto.PubKey
	Controller string
}

func (a *RegisterAssetInfo) Data() []byte {
	return []byte{0}
}

//implement PayloadInfo define TransferAssetInfo
type TransferAssetInfo struct {
}

func (a *TransferAssetInfo) Data() []byte {
	return []byte{0}
}

func TransPayloadToHex(p Payload) PayloadInfo {
	switch object := p.(type) {
	case *payload.BookKeeping:
	case *payload.IssueAsset:
	case *payload.TransferAsset:
	case *payload.DeployCode:
		obj := new(DeployCodeInfo)
		obj.Code.Code = ToHexString(object.Code.Code)
		obj.Code.ParameterTypes = ToHexString(ContractParameterTypeToByte(object.Code.ParameterTypes))
		obj.Code.ReturnTypes = ToHexString(ContractParameterTypeToByte(object.Code.ReturnTypes))
		obj.Name = object.Name
		obj.CodeVersion = object.CodeVersion
		obj.Author = object.Author
		obj.Email = object.Email
		obj.Description = object.Description
		return obj
	case *payload.RegisterAsset:
		obj := new(RegisterAssetInfo)
		obj.Asset = object.Asset
		obj.Amount = object.Amount
		obj.Issuer = object.Issuer
		obj.Controller = ToHexString(object.Controller.ToArray())
		return obj
	}
	return nil
}

