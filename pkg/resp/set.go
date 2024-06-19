package resp

import (
	"io"
	"math/rand/v2"
	"sync/atomic"
)

type Set struct {
	data hashedArray[Value]
	hash uint64
	attr *Attribute
}

var (
	EmptySet = &Set{
		data: makeHashedArray[Value](0),
	}

	emptySetHash = rand.Uint64()
)

// compile-time checks for interface implementation
var _ interface {
	Counted
	Hasher
	Collection
} = (*Set)(nil)

func readSet(r *Reader) (*Set, error) {
	resLen, err := r.readLen()
	if err != nil {
		return nil, err
	}
	res := &Set{
		data: makeHashedArray[Value](resLen),
	}
	for i := 0; i < resLen; i++ {
		v, err := r.Next()
		if err != nil {
			return nil, err
		}
		res.data.put(v)
	}
	return res, nil
}

func MakeSet(v ...Value) *Set {
	res := &Set{
		data: makeHashedArray[Value](len(v)),
	}
	for _, e := range v {
		res.data.put(e)
	}
	return res
}

func (*Set) Tag() Tag {
	return SetTag
}

func (s *Set) Marshal(w io.Writer) error {
	if _, err := w.Write([]byte{byte(s.Tag())}); err != nil {
		return err
	}
	return s.data.marshal(w)
}

func (s *Set) Elements() Values {
	res := make(Values, 0, len(s.data))
	for _, e := range s.data {
	next:
		res = append(res, e.ref)
		if e.next != nil {
			e = e.next
			goto next
		}
	}
	return res
}

func (s *Set) Contains(v Value) bool {
	bucket := Hash(v) % uint64(len(s.data))
	if e := s.data[bucket]; e != nil {
		return e.contains(v) != nil
	}
	return false
}

func (s *Set) ForEach(fn func(Value) error) error {
	return s.data.forEach(fn)
}

func (s *Set) Count() int {
	return s.data.count()
}

func (s *Set) Equal(v Value) bool {
	if v, ok := v.(*Set); ok {
		return s.data.equal(v.data)
	}
	return false
}

func (s *Set) Hash() uint64 {
	if h := atomic.LoadUint64(&s.hash); h != 0 {
		return h
	}
	h := emptySetHash ^ s.data.hash()
	atomic.StoreUint64(&s.hash, h)
	return h
}

func (s *Set) WithAttribute(attr *Attribute) Value {
	res := *s
	res.attr = attr
	return &res
}

func (s *Set) Attribute() *Attribute {
	return s.attr
}
