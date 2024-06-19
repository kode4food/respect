package resp

import "io"

type (
	Mapped interface {
		Counted
		Hasher
		Get(key Value) (Value, bool)
		ForEach(fn func(Value, Value) error) error

		put(key, val Value)
	}

	mapped struct {
		data hashedArray[*mappedPair]
	}

	mappedPair struct {
		key   Value
		value Value
	}
)

func newMapped(size int) mapped {
	return mapped{
		data: makeHashedArray[*mappedPair](size),
	}
}

func readMapped(r *Reader) (*mapped, error) {
	resLen, err := r.readLen()
	if err != nil {
		return nil, err
	}
	res := &mapped{
		data: makeHashedArray[*mappedPair](resLen),
	}
	for i := 0; i < resLen; i++ {
		key, err := r.Next()
		if err != nil {
			return nil, err
		}
		val, err := r.Next()
		if err != nil {
			return nil, err
		}
		res.put(key, val)
	}
	return res, nil
}

func makeFromMap[K MakeMappedKey, V Value](m *mapped, p map[K]V) {
	for k, v := range p {
		m.put(k, v)
	}
}

func makeFromPairs(m *mapped, pairs ...[2]Value) {
	for _, p := range pairs {
		m.put(p[0], p[1])
	}
}

func (m *mapped) marshal(t Tag, w io.Writer) error {
	if _, err := w.Write([]byte{byte(t)}); err != nil {
		return err
	}
	return m.data.marshal(w)
}

func (m *mapped) Get(key Value) (Value, bool) {
	bucket := Hash(key) % uint64(len(m.data))
	for e := m.data[bucket]; e != nil; e = e.next {
		if e.ref.key.Equal(key) {
			return e.ref.value, true
		}
	}
	return nil, false
}

func (m *mapped) ForEach(fn func(Value, Value) error) error {
	return m.data.forEach(func(p *mappedPair) error {
		return fn(p.key, p.value)
	})
}

func (m *mapped) Count() int {
	return m.data.count()
}

func (m *mapped) put(key, val Value) {
	bucket := Hash(key) % uint64(len(m.data))
	e := m.data[bucket]
	if e == nil {
		m.data[bucket] = &hashedEntry[*mappedPair]{
			ref: &mappedPair{
				key: key, value: val,
			},
		}
		return
	}
	for i := e; i != nil; i = i.next {
		if i.ref.key.Equal(key) {
			i.ref.value = val
			return
		}
	}
	m.data[bucket] = &hashedEntry[*mappedPair]{
		ref: &mappedPair{
			key:   key,
			value: val,
		},
		next: e,
	}
}

func (*mappedPair) Tag() Tag {
	return 0
}

func (p *mappedPair) Equal(v Value) bool {
	if v, ok := v.(*mappedPair); ok {
		return p.key.Equal(v.key) && p.value.Equal(v.value)
	}
	return false
}

func (p *mappedPair) Marshal(w io.Writer) error {
	if err := p.key.Marshal(w); err != nil {
		return err
	}
	return p.value.Marshal(w)
}

func (p *mappedPair) Hash() uint64 {
	return Hash(p.key) ^ Hash(p.value)
}
