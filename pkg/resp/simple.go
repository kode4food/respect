package resp

import (
	"fmt"
	"io"
)

type (
	Null    struct{}
	Boolean bool
)

// Error messages
const (
	ErrInvalidBoolean = "ERR invalid boolean: %s"
)

var (
	NullValue = Null{}
	True      = Boolean(true)
	False     = Boolean(false)

	trueBytes  = []byte{'t'}
	falseBytes = []byte{'f'}
)

// compile-time checks for interface implementation
var (
	_ Value = Null{}
	_ Value = Boolean(false)
)

func readNull(r *Reader) (Null, error) {
	data, err := r.readSimple()
	if err != nil {
		return NullValue, err
	}
	if len(data) != 0 {
		return NullValue, fmt.Errorf(ErrInvalidLength, len(data))
	}
	return NullValue, nil
}

func (Null) Tag() Tag {
	return NullTag
}

func (n Null) Marshal(w io.Writer) error {
	return writeSimple(n.Tag(), nil, w)
}

func (Null) Equal(v Value) bool {
	_, ok := v.(Null)
	return ok
}

func (Null) String() string {
	return "(null)"
}

func readBoolean(r *Reader) (Value, error) {
	data, err := r.readSimple()
	if err != nil {
		return nil, err
	}
	if len(data) != 1 {
		return nil, fmt.Errorf(ErrInvalidLength, len(data))
	}
	switch data[0] {
	case 't':
		return True, nil
	case 'f':
		return False, nil
	default:
		return nil, fmt.Errorf(ErrInvalidBoolean, string(data))
	}
}

func (Boolean) Tag() Tag {
	return BooleanTag
}

func (b Boolean) Marshal(w io.Writer) error {
	return writeSimple(b.Tag(), b.trueOrFalseBytes(), w)
}

func (b Boolean) Bool() bool {
	return bool(b)
}

func (b Boolean) Equal(v Value) bool {
	if v, ok := v.(Boolean); ok {
		return b == v
	}
	return false
}

func (b Boolean) String() string {
	if b {
		return "(true)"
	}
	return "(false)"
}

func (b Boolean) trueOrFalseBytes() []byte {
	if b {
		return trueBytes
	}
	return falseBytes
}
