package errors

import "errors"

var (
	ErrAssetNameInvalid = errors.New("asset name invalid")
	ErrAssetPrecisionInvalid = errors.New("asset precision invalid")
	ErrAssetAmountInvalid = errors.New("asset amount invalid")
)
