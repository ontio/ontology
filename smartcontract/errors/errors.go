package errors

import "errors"

var (
	ErrAssetNameInvalid = errors.New("asset name invalid, too long")
	ErrAssetPrecisionInvalid = errors.New("asset precision invalid")
	ErrAssetAmountInvalid = errors.New("asset amount invalid")
	ErrAssetCheckOwnerInvalid = errors.New("asset owner invalid")
)
