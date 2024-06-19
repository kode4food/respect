package resp

import "slices"

type (
	Collection interface {
		ForEach(func(Value) error) error
		Contains(Value) bool
		Elements() Values
	}

	Values []Value
)

// compile-time checks for interface implementation
var _ Collection = Values{}

func (v Values) ForEach(fn func(Value) error) error {
	for _, e := range v {
		if err := fn(e); err != nil {
			return err
		}
	}
	return nil
}

func (v Values) Elements() Values {
	return v
}

func (v Values) Contains(e Value) bool {
	for _, l := range v {
		if l.Equal(e) {
			return true
		}
	}
	return false
}

func (v Values) Count() int {
	return len(v)
}

func (v Values) Equal(other Values) bool {
	if len(v) != len(other) {
		return false
	}
	for i, l := range v {
		if !l.Equal(other[i]) {
			return false
		}
	}
	return true
}

func (v Values) Hash() uint64 {
	var h uint64
	for _, e := range v {
		h ^= Hash(e)
	}
	return h
}

func (v Values) Sort() Values {
	return v.SortWith(Compare)
}

func (v Values) SortWith(cmp func(Value, Value) int) Values {
	if len(v) < 2 {
		return v
	}
	res := make(Values, len(v))
	copy(res, v)
	slices.SortFunc(res, cmp)
	return res
}
