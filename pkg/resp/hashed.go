package resp

import "io"

type (
	hashedArray[T Value] []*hashedEntry[T]

	hashedEntry[T Value] struct {
		ref  T
		next *hashedEntry[T]
	}
)

func makeHashedArray[T Value](size int) hashedArray[T] {
	return make([]*hashedEntry[T], size)
}

func (a hashedArray[T]) count() int {
	res := 0
	for _, e := range a {
		res += e.count()
	}
	return res
}

func (a hashedArray[T]) put(v T) {
	bucket := Hash(v) % uint64(len(a))
	e := a[bucket]
	if e == nil {
		a[bucket] = &hashedEntry[T]{ref: v}
		return
	}
	if e.contains(v) != nil {
		return
	}
	a[bucket] = &hashedEntry[T]{
		ref:  v,
		next: e,
	}
}

func (a hashedArray[T]) forEach(fn func(T) error) error {
	for _, e := range a {
	next:
		if e == nil {
			continue
		}
		if err := fn(e.ref); err != nil {
			return err
		}
		e = e.next
		goto next
	}
	return nil
}

func (a hashedArray[T]) marshal(w io.Writer) error {
	if err := writeInt(len(a), w); err != nil {
		return err
	}
	return a.forEach(func(v T) error {
		if err := v.Marshal(w); err != nil {
			return err
		}
		return nil
	})
}

func (a hashedArray[T]) equal(other hashedArray[T]) bool {
	if len(a) != len(other) {
		return false
	}
	for i, e := range a {
		oe := other[i]
		if e == nil {
			if oe != nil {
				return false
			}
			continue
		}
	next:
		if oe.contains(e.ref) == nil {
			return false
		}
		if e.next != nil {
			e = e.next
			goto next
		}
	}
	return true
}

func (a hashedArray[T]) hash() uint64 {
	var res uint64
	for _, e := range a {
	next:
		if e == nil {
			continue
		}
		res ^= Hash(e.ref)
		if e.next != nil {
			e = e.next
			goto next
		}
	}
	return res
}

func (e *hashedEntry[T]) contains(v T) *hashedEntry[T] {
	for i := e; i != nil; i = i.next {
		if i.ref.Equal(v) {
			return i
		}
	}
	return nil
}

func (e *hashedEntry[T]) count() int {
	res := 0
	for i := e; i != nil; i = i.next {
		res++
	}
	return res
}
