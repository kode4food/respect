package resp

import (
	"fmt"
	"io"
	"math/big"
	"strconv"
)

type (
	Integer   int64
	Double    float64
	BigNumber big.Int
)

// Error messages
const (
	ErrInvalidBigNumber = "ERR invalid big number: %s"
)

var (
	ZeroInteger   = Integer(0)
	ZeroDouble    = Double(0)
	ZeroBigNumber = (*BigNumber)(new(big.Int).SetInt64(0))
)

// compile-time checks for interface implementation
var (
	_ Value = Integer(0)
	_ Value = Double(0)
	_ Value = (*BigNumber)(nil)
)

func readInteger(r *Reader) (Integer, error) {
	i, err := r.readInt64()
	if err != nil {
		return ZeroInteger, err
	}
	return Integer(i), nil
}

func (Integer) Tag() Tag {
	return IntegerTag
}

func (i Integer) Marshal(w io.Writer) error {
	return writeSimple(i.Tag(), []byte(i.String()), w)
}

func (i Integer) Equal(v Value) bool {
	if v, ok := v.(Integer); ok {
		return i == v
	}
	return false
}

func (i Integer) String() string {
	return strconv.FormatInt(int64(i), 10)
}

func readDouble(r *Reader) (Double, error) {
	data, err := r.readSimple()
	if err != nil {
		return ZeroDouble, err
	}
	f, err := strconv.ParseFloat(string(data), 64)
	if err != nil {
		return ZeroDouble, err
	}
	return Double(f), nil
}

func (Double) Tag() Tag {
	return DoubleTag
}

func (d Double) Marshal(w io.Writer) error {
	return writeSimple(d.Tag(), []byte(d.String()), w)
}

func (d Double) Equal(v Value) bool {
	if v, ok := v.(Double); ok {
		return d == v
	}
	return false
}

func (d Double) String() string {
	return strconv.FormatFloat(float64(d), 'f', -1, 64)
}

func MakeBigNumber(s string) (*BigNumber, error) {
	b, ok := new(big.Int).SetString(s, 10)
	if !ok {
		return ZeroBigNumber, fmt.Errorf(ErrInvalidBigNumber, s)
	}
	return (*BigNumber)(b), nil
}

func readBigNumber(r *Reader) (*BigNumber, error) {
	data, err := r.readSimple()
	if err != nil {
		return ZeroBigNumber, err
	}
	return MakeBigNumber(string(data))
}

func (*BigNumber) Tag() Tag {
	return BigNumberTag
}

func (b *BigNumber) Marshal(w io.Writer) error {
	return writeSimple(b.Tag(), []byte(b.String()), w)
}

func (b *BigNumber) Equal(v Value) bool {
	if v, ok := v.(*BigNumber); ok {
		return (*big.Int)(b).Cmp((*big.Int)(v)) == 0
	}
	return false
}

func (b *BigNumber) String() string {
	return (*big.Int)(b).String()
}
