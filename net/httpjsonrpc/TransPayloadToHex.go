package httpjsonrpc

import (
	. "DNA/common"
	"DNA/core/asset"
	. "DNA/core/contract"
	. "DNA/core/transaction"
	"DNA/core/transaction/payload"
	"bytes"
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

type IssuerInfo struct {
	X, Y string
}

//implement PayloadInfo define RegisterAssetInfo
type RegisterAssetInfo struct {
	Asset      *asset.Asset
	Amount     Fixed64
	Issuer     IssuerInfo
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

type RecordInfo struct {
	RecordType string
	RecordData string
}

func (a *RecordInfo) Data() []byte {
	return []byte{0}
}

type PrivacyPayloadInfo struct {
	PayloadType uint8
	Payload     string
	EncryptType uint8
	EncryptAttr string
}

func (a *PrivacyPayloadInfo) Data() []byte {
	return []byte{0}
}

func TransPayloadToHex(p Payload) PayloadInfo {
	switch object := p.(type) {
	case *payload.BookKeeping:
	case *payload.BookKeeper:
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
		obj.Issuer.X = object.Issuer.X.String()
		obj.Issuer.Y = object.Issuer.Y.String()
		obj.Controller = ToHexString(object.Controller.ToArray())
		return obj
	case *payload.Record:
		obj := new(RecordInfo)
		obj.RecordType = object.RecordType
		obj.RecordData = ToHexString(object.RecordData)
		return obj
	case *payload.PrivacyPayload:
		obj := new(PrivacyPayloadInfo)
		obj.PayloadType = uint8(object.PayloadType)
		obj.Payload = ToHexString(object.Payload)
		obj.EncryptType = uint8(object.EncryptType)
		bytesBuffer := bytes.NewBuffer([]byte{})
		object.EncryptAttr.Serialize(bytesBuffer)
		obj.EncryptAttr = ToHexString(bytesBuffer.Bytes())
		return obj

	}
	return nil
}
