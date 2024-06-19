package resp

import (
	"bytes"
	"fmt"
	"io"
)

type (
	String interface {
		Value
		fmt.Stringer
	}

	SimpleString string
	BulkString   string
)

var (
	EmptySimpleString = SimpleString("")
	EmptyBulkString   = BulkString("")
)

// Error messages
const (
	ErrInvalidEncoding = "invalid encoding: %s"
)

// compile-time checks for interface implementation
var (
	_ String = SimpleString("")
	_ String = BulkString("")
)

func MakeString(s string) String {
	if bytes.Contains([]byte(s), NewLine) {
		return BulkString(s)
	}
	return SimpleString(s)
}

func readSimpleString(r *Reader) (SimpleString, error) {
	data, err := r.readSimple()
	if err != nil {
		return EmptySimpleString, err
	}
	return SimpleString(data), nil
}

func (SimpleString) Tag() Tag {
	return SimpleStringTag
}

func (s SimpleString) Marshal(w io.Writer) error {
	return writeSimple(s.Tag(), []byte(s), w)
}

func (s SimpleString) Equal(v Value) bool {
	if v, ok := v.(SimpleString); ok {
		return s == v
	}
	return false
}

func (s SimpleString) String() string {
	return string(s)
}

func readBulkString(r *Reader) (BulkString, error) {
	data, err := r.readBulk()
	if err != nil {
		return EmptyBulkString, err
	}
	return BulkString(data), nil
}

func (BulkString) Tag() Tag {
	return BulkStringTag
}

func (s BulkString) Marshal(w io.Writer) error {
	return writeBulk(s.Tag(), []byte(s), w)
}

func (s BulkString) Equal(v Value) bool {
	if v, ok := v.(BulkString); ok {
		return s == v
	}
	return false
}

func (s BulkString) String() string {
	return string(s)
}
