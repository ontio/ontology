package utils

import (
	"fmt"
	"math"
	"math/big"
)

const (
	PRECISION_ONG = 9
	PRECISION_ONT = 0
)

//FormatAssetAmount return asset amount multiplied by math.Pow10(precision) to raw float string
//For example 1000000000123456789 => 1000000000.123456789
func FormatAssetAmount(amount uint64, precision byte) string {
	if precision == 0 {
		return fmt.Sprintf("%d", amount)
	}
	divisor := math.Pow10(int(precision))
	intPart := amount / uint64(divisor)
	fracPart := amount - intPart*uint64(divisor)
	if fracPart == 0 {
		return fmt.Sprintf("%d", intPart)
	}
	bf := new(big.Float).SetUint64(fracPart)
	bf.Quo(bf, new(big.Float).SetFloat64(math.Pow10(int(precision))))
	bf.Add(bf, new(big.Float).SetUint64(intPart))
	return bf.Text('f', -1)
}

//ParseAssetAmount return raw float string to uint64 multiplied by math.Pow10(precision)
//For example 1000000000.123456789 => 1000000000123456789
func ParseAssetAmount(rawAmount string, precision byte) uint64 {
	bf, ok := new(big.Float).SetString(rawAmount)
	if !ok {
		return 0
	}
	bf.Mul(bf, new(big.Float).SetFloat64(math.Pow10(int(precision))))
	amount, _ := bf.Uint64()
	return amount
}

func FormatOng(amount uint64) string {
	return FormatAssetAmount(amount, PRECISION_ONG)
}

func ParseOng(rawAmount string) uint64 {
	return ParseAssetAmount(rawAmount, PRECISION_ONG)
}

func FormatOnt(amount uint64) string {
	return FormatAssetAmount(amount, PRECISION_ONT)
}

func ParseOnt(rawAmount string) uint64 {
	return ParseAssetAmount(rawAmount, PRECISION_ONT)
}
