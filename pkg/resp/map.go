package resp

import (
	"io"
	"math/rand/v2"
	"sync/atomic"
)

type (
	MakeMappedKey interface {
		comparable
		Value
	}

	Map struct {
		mapped
		hash uint64
	}
)

var (
	EmptyMap = &Map{mapped: newMapped(0)}

	emptyMapHash = rand.Uint64()
)

// compile-time check for interface implementation
var _ Mapped = (*Map)(nil)

func readMap(r *Reader) (*Map, error) {
	m, err := readMapped(r)
	if err != nil {
		return nil, err
	}
	return &Map{mapped: *m}, nil
}

func MakeMap[K MakeMappedKey, V Value](m map[K]V) *Map {
	res := &Map{mapped: newMapped(len(m))}
	makeFromMap(&res.mapped, m)
	return res
}

func MakeMapFromPairs(pairs ...[2]Value) *Map {
	res := &Map{mapped: newMapped(len(pairs))}
	makeFromPairs(&res.mapped, pairs...)
	return res
}

func (*Map) Tag() Tag {
	return MapTag
}

func (m *Map) Marshal(w io.Writer) error {
	return m.marshal(MapTag, w)
}

func (m *Map) Equal(v Value) bool {
	if v, ok := v.(*Map); ok {
		return m.data.equal(v.data)
	}
	return false
}

func (m *Map) Hash() uint64 {
	if h := atomic.LoadUint64(&m.hash); h != 0 {
		return h
	}
	h := emptyMapHash ^ m.data.hash()
	atomic.StoreUint64(&m.hash, h)
	return h
}
