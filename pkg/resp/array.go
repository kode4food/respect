package resp

import (
	"io"
	"math/rand/v2"
	"sync/atomic"
)

type Array struct {
	Values
	hash uint64
}

var (
	EmptyArray = &Array{Values: Values{}}

	emptyArrayHash = rand.Uint64()
)

// compile-time checks for interface implementation
var _ interface {
	Counted
	Hasher
	Collection
} = (*Array)(nil)

func readArray(r *Reader) (*Array, error) {
	res, err := r.readValues()
	if err != nil {
		return EmptyArray, err
	}
	return MakeArray(res...), nil
}

func MakeArray(v ...Value) *Array {
	return &Array{Values: v}
}

func (*Array) Tag() Tag {
	return ArrayTag
}

func (a *Array) Marshal(w io.Writer) error {
	return writeValues(a.Tag(), a.Values, w)
}

func (a *Array) Equal(v Value) bool {
	if v, ok := v.(*Array); ok {
		return a.Values.Equal(v.Values)
	}
	return false
}

func (a *Array) Hash() uint64 {
	if h := atomic.LoadUint64(&a.hash); h != 0 {
		return h
	}
	h := emptyArrayHash ^ a.Values.Hash()
	atomic.StoreUint64(&a.hash, h)
	return h
}
