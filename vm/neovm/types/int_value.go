package types

import (
	"math/big"

	"github.com/JohnCGriffin/overflow"
	"github.com/ontio/ontology/common"
	"github.com/ontio/ontology/vm/neovm/constants"
	"github.com/ontio/ontology/vm/neovm/errors"
	"math"
)

type IntValue struct {
	isbig   bool
	integer int64
	bigint  *big.Int
}

func (self IntValue) Rsh(other IntValue) (result IntValue, err error) {
	var val uint64
	if other.isbig == false {
		if other.integer < 0 {
			err = errors.ERR_SHIFT_BY_NEG
			return
		}
		val = uint64(other.integer)
	} else {
		if other.bigint.IsUint64() == false {
			return IntValue{}, errors.ERR_SHIFT_BY_NEG
		}
		val = other.bigint.Uint64()
	}

	if val > constants.MAX_INT_SIZE*8 {
		// IntValue is enforced to not exceed this size, so return 0 directly
		return
	}

	left := big.NewInt(self.integer)
	if self.isbig {
		left.Set(self.bigint)
	}

	left.Rsh(left, uint(val))

	return IntValFromBigInt(left)
}

func (self IntValue) Lsh(other IntValue) (result IntValue, err error) {
	var val uint64
	if other.isbig == false {
		if other.integer < 0 {
			err = errors.ERR_SHIFT_BY_NEG
			return
		}
		val = uint64(other.integer)
	} else {
		if other.bigint.IsUint64() == false {
			return IntValue{}, errors.ERR_SHIFT_BY_NEG
		}
		val = other.bigint.Uint64()
	}

	if val > constants.MAX_INT_SIZE*8 {
		err = errors.ERR_OVER_MAX_BIGINTEGER_SIZE
		return
	}

	left := big.NewInt(self.integer)
	if self.isbig {
		left.Set(self.bigint)
	}

	left.Lsh(left, uint(val))

	return IntValFromBigInt(left)
}

func IntValFromBigInt(val *big.Int) (result IntValue, err error) {
	if val == nil {
		err = errors.ERR_BAD_VALUE
	}
	// todo : check compatibility
	if len(val.Bytes()) > constants.MAX_INT_SIZE {
		err = errors.ERR_OVER_MAX_BIGINTEGER_SIZE
	}

	if val.IsInt64() {
		result.isbig = false
		result.integer = val.Int64()
	} else {
		result.isbig = true
		result.bigint = val
	}

	return
}

func IntValFromNeoBytes(val []byte) (IntValue, error) {
	value := common.BigIntFromNeoBytes(val)
	return IntValFromBigInt(value)
}

func (self *IntValue) ToNeoBytes() []byte {
	val := self.bigint
	if self.isbig == false {
		val = big.NewInt(self.integer)
	}
	value := common.BigIntToNeoBytes(val)
	return value
}

func IntValFromInt(val int64) IntValue {
	return IntValue{isbig: false, integer: val}
}

func (self *IntValue) IsZero() bool {
	if self.isbig {
		return self.bigint.Sign() == 0
	} else {
		return self.integer == 0
	}
}

func (self *IntValue) Sign() int {
	if self.isbig {
		return self.bigint.Sign()
	} else {
		if self.integer < 0 {
			return -1
		} else if self.integer == 0 {
			return 0
		} else {
			return 1
		}
	}
}

func (self IntValue) Max(other IntValue) (IntValue, error) {
	return self.intOp(other, func(a, b int64) (int64, bool) {
		if a < b {
			return b, true
		}
		return a, true
	}, func(a, b *big.Int) (IntValue, error) {
		result := new(big.Int)
		c := a.Cmp(b)
		if c <= 0 {
			result.Set(b)
		} else {
			result.Set(a)
		}
		return IntValFromBigInt(result)
	})
}

func (self IntValue) Min(other IntValue) (IntValue, error) {
	return self.intOp(other, func(a, b int64) (int64, bool) {
		if a < b {
			return a, true
		}
		return b, true
	}, func(a, b *big.Int) (IntValue, error) {
		result := new(big.Int)
		c := a.Cmp(b)
		if c <= 0 {
			result.Set(a)
		} else {
			result.Set(b)
		}
		return IntValFromBigInt(result)
	})
}

func (self IntValue) Xor(other IntValue) (IntValue, error) {
	return self.intOp(other, func(a, b int64) (int64, bool) {
		return a ^ b, true
	}, func(a, b *big.Int) (IntValue, error) {
		return IntValFromBigInt(new(big.Int).And(a, b))
	})
}

func (self IntValue) And(other IntValue) (IntValue, error) {
	return self.intOp(other, func(a, b int64) (int64, bool) {
		return a & b, true
	}, func(a, b *big.Int) (IntValue, error) {
		return IntValFromBigInt(new(big.Int).And(a, b))
	})
}

func (self IntValue) Or(other IntValue) (IntValue, error) {
	return self.intOp(other, func(a, b int64) (int64, bool) {
		return a | b, true
	}, func(a, b *big.Int) (IntValue, error) {
		return IntValFromBigInt(new(big.Int).Or(a, b))
	})
}

func (self IntValue) Cmp(other IntValue) int {
	if self.isbig == false && other.isbig == false {
		if self.integer < other.integer {
			return -1
		} else if self.integer == other.integer {
			return 0
		} else {
			return 1
		}
	}
	var left, right *big.Int
	if self.isbig == false {
		left = big.NewInt(self.integer)
	} else {
		left = self.bigint
	}
	if other.isbig == false {
		right = big.NewInt(other.integer)
	} else {
		right = other.bigint
	}

	return left.Cmp(right)
}

func (self IntValue) Not() (val IntValue) {
	if self.isbig {
		val.isbig = true
		val.bigint = big.NewInt(0)
		val.bigint.Not(self.bigint)
	} else {
		val.integer = ^self.integer
	}
	return
}

func (self IntValue) Abs() (val IntValue) {
	if self.isbig {
		val.isbig = true
		val.bigint = big.NewInt(0)
		val.bigint.Abs(self.bigint)
	} else {
		if self.integer == math.MinInt64 {
			val.isbig = true
			val.bigint = big.NewInt(self.integer)
			val.bigint.Abs(val.bigint)
		} else {
			val.integer = -self.integer
		}
	}
	return
}

// todo: check negative value with big.Int
func (self IntValue) Mod(other IntValue) (IntValue, error) {
	if other.IsZero() {
		return IntValue{}, errors.ERR_DIV_MOD_BY_ZERO
	}
	return self.intOp(other, func(a, b int64) (int64, bool) {
		return a % b, true
	}, func(a, b *big.Int) (IntValue, error) {
		return IntValFromBigInt(new(big.Int).Rem(a, b))
	})
}

// todo: check negative value with big.Int
func (self IntValue) Div(other IntValue) (IntValue, error) {
	if other.IsZero() {
		return IntValue{}, errors.ERR_DIV_MOD_BY_ZERO
	}
	return self.intOp(other, func(a, b int64) (int64, bool) {
		return a / b, true
	}, func(a, b *big.Int) (IntValue, error) {
		return IntValFromBigInt(new(big.Int).Quo(a, b))
	})
}

func (self IntValue) Mul(other IntValue) (IntValue, error) {
	return self.intOp(other, overflow.Mul64, func(a, b *big.Int) (IntValue, error) {
		return IntValFromBigInt(new(big.Int).Mul(a, b))
	})
}

func (self IntValue) Add(other IntValue) (IntValue, error) {
	return self.intOp(other, overflow.Add64, func(a, b *big.Int) (IntValue, error) {
		return IntValFromBigInt(new(big.Int).Add(a, b))
	})
}

func (self IntValue) Sub(other IntValue) (IntValue, error) {
	return self.intOp(other, overflow.Sub64, func(a, b *big.Int) (IntValue, error) {
		return IntValFromBigInt(new(big.Int).Sub(a, b))
	})
}

type overflowFn func(a, b int64) (result int64, ok bool)
type bigintFn func(a, b *big.Int) (IntValue, error)

func (self IntValue) intOp(other IntValue, littleintFunc overflowFn, bigintFunc bigintFn) (IntValue, error) {
	if self.isbig == false && other.isbig == false {
		val, ok := littleintFunc(self.integer, other.integer)
		if ok {
			return IntValFromInt(val), nil
		}
	}
	var left, right *big.Int
	if self.isbig == false {
		left = big.NewInt(self.integer)
	} else {
		left = self.bigint
	}
	if other.isbig == false {
		right = big.NewInt(other.integer)
	} else {
		right = other.bigint
	}

	return bigintFunc(left, right)
}
