package resp

import (
	"io"
	"math/rand/v2"
	"sync/atomic"
)

type Attribute struct {
	mapped
	hash uint64
}

// compile-time check for interface implementation
var _ Mapped = (*Attribute)(nil)

var (
	EmptyAttribute = &Attribute{mapped: newMapped(0)}

	emptyAttributeHash = rand.Uint64()
)

func readAttribute(r *Reader) (*Attribute, error) {
	m, err := readMapped(r)
	if err != nil {
		return nil, err
	}
	return &Attribute{mapped: *m}, nil
}

func MakeAttribute[K MakeMappedKey, V Value](m map[K]V) *Attribute {
	res := &Attribute{mapped: newMapped(len(m))}
	makeFromMap(&res.mapped, m)
	return res
}

func MakeAttributeFromPairs(pairs ...[2]Value) *Attribute {
	res := &Attribute{mapped: newMapped(len(pairs))}
	makeFromPairs(&res.mapped, pairs...)
	return res
}

func (*Attribute) Tag() Tag {
	return AttributeTag
}

func (a *Attribute) Marshal(w io.Writer) error {
	return a.marshal(AttributeTag, w)
}

func (a *Attribute) Equal(v Value) bool {
	if v, ok := v.(*Map); ok {
		return a.data.equal(v.data)
	}
	return false
}

func (a *Attribute) Hash() uint64 {
	if h := atomic.LoadUint64(&a.hash); h != 0 {
		return h
	}
	h := emptyAttributeHash ^ a.data.hash()
	atomic.StoreUint64(&a.hash, h)
	return h
}
