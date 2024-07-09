package resp

import (
	"bufio"
	"cmp"
	"fmt"
	"hash/maphash"
	"io"
	"math/big"
	"strings"
)

type (
	// Value represents a single RESP value
	Value interface {
		// Tag returns the type of the Value
		Tag() Tag

		// Marshal writes a readable representation of the Value to a provided
		// io.Writer. Implementations assume Writer is in some way buffered,
		// and may make multiple calls to Write to avoid building unnecessary
		// slices in memory
		Marshal(io.Writer) error

		// Equal compares the Value to another Value, returning true if Equal
		Equal(Value) bool
	}

	Hasher interface {
		Value
		Hash() uint64
	}

	// Counted represents a RESP value that has a count of elements
	Counted interface {
		Value
		Count() int
	}

	TopLevelOnly interface {
		Value
		topLevel()
	}
)

const (
	OK = SimpleString("OK")
)

var seed = maphash.MakeSeed()

func Hash(v Value) uint64 {
	switch t := v.(type) {
	case Hasher:
		return t.Hash()
	case fmt.Stringer:
		return hashStringer(v.Tag(), t)
	default:
		return hashValue(v)
	}
}

func hashStringer(t Tag, v fmt.Stringer) uint64 {
	h := maphash.Hash{}
	h.SetSeed(seed)
	_ = h.WriteByte(byte(t))
	_, _ = h.WriteString(v.String())
	return h.Sum64()
}

func hashValue(v Value) uint64 {
	h := maphash.Hash{}
	h.SetSeed(seed)
	_ = h.WriteByte(byte(v.Tag()))
	_ = v.Marshal(&h)
	return h.Sum64()
}

func ReadString(s string, opts ...ReaderOption) (Value, error) {
	sr := strings.NewReader(s)
	b := bufio.NewReader(sr)
	r := NewReader(b, opts...)
	return r.Next()
}

func ToString(v Value) string {
	var sb strings.Builder
	_ = v.Marshal(&sb)
	return sb.String()
}

func Compare(l, r Value) int {
	if l == r {
		return 0
	}
	if l.Tag() != r.Tag() {
		return cmp.Compare(l.Tag(), r.Tag())
	}
	switch l := l.(type) {
	case Integer:
		return cmpOrdered(l, r.(Integer))
	case Double:
		return cmpOrdered(l, r.(Double))
	case *BigNumber:
		return cmpBigNumber(l, r.(*BigNumber))
	case Boolean:
		return cmpBoolean(l, r.(Boolean))
	case fmt.Stringer:
		return cmpStringer(l, r.(fmt.Stringer))
	default:
		return cmpMarshaled(l, r)
	}
}

func cmpOrdered[T cmp.Ordered](l, r T) int {
	return cmp.Compare(l, r)
}

func cmpBigNumber(l, r *BigNumber) int {
	return (*big.Int)(l).Cmp((*big.Int)(r))
}

func cmpBoolean(l, r Boolean) int {
	if l == r {
		return 0
	}
	if !l {
		return -1
	}
	return 1
}

func cmpStringer(l, r fmt.Stringer) int {
	return cmp.Compare(l.String(), r.String())
}

func cmpMarshaled(l, r Value) int {
	ls := ToString(l)
	rs := ToString(r)
	return cmp.Compare(ls, rs)
}
