package resp

import (
	"io"
	"math/rand/v2"
	"sync/atomic"
)

type Push struct {
	Values
	hash uint64
}

var (
	EmptyPush = &Push{Values: Values{}}

	emptyPushHash = rand.Uint64()
)

// compile-time checks for interface implementation
var _ interface {
	Counted
	Hasher
	Collection
	TopLevelOnly
} = (*Push)(nil)

func (*Push) topLevel() {}

func readPush(r *Reader) (*Push, error) {
	res, err := r.readValues()
	if err != nil {
		return EmptyPush, err
	}
	return MakePush(res...), nil
}

func MakePush(v ...Value) *Push {
	return &Push{Values: v}
}

func (*Push) Tag() Tag {
	return PushTag
}

func (p *Push) Marshal(w io.Writer) error {
	return writeValues(p.Tag(), p.Values, w)
}

func (p *Push) Equal(v Value) bool {
	if v, ok := v.(*Push); ok {
		return p.Values.Equal(v.Values)
	}
	return false
}

func (p *Push) Hash() uint64 {
	if h := atomic.LoadUint64(&p.hash); h != 0 {
		return h
	}
	h := emptyPushHash ^ p.Values.Hash()
	atomic.StoreUint64(&p.hash, h)
	return h
}
