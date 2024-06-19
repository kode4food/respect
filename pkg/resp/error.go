package resp

import (
	"bytes"
	"fmt"
	"io"
	"regexp"
)

type (
	Error interface {
		Value
		error()
		Prefix() string
		fmt.Stringer
		error
	}

	errorStr string

	SimpleError struct{ errorStr }
	BulkError   struct{ errorStr }
)

var prefixedError = regexp.MustCompile(`^([A-Z]{2,})\s+(.+)$`)

// compile-time checks for interface implementation
var (
	_ Error = (*SimpleError)(nil)
	_ Error = (*BulkError)(nil)
)

func (*SimpleError) error()  {}
func (*SimpleError) simple() {}
func (*BulkError) error()    {}
func (*BulkError) bulk()     {}

func readSimpleError(r *Reader) (*SimpleError, error) {
	data, err := r.readSimple()
	if err != nil {
		return nil, err
	}
	return MakeSimpleError(string(data)), nil
}

// MakeError creates an Error from a string and optional arguments. If the
// resulting string contains a CR/LF sequence, a BulkError will be created,
// otherwise a SimpleError will be created.
func MakeError(s string, args ...any) Error {
	e := fmt.Sprintf(s, args...)
	if bytes.Contains([]byte(e), NewLine) {
		return MakeBulkError(e)
	}
	return MakeSimpleError(e)
}

func MakeSimpleError(s string) *SimpleError {
	return &SimpleError{errorStr(s)}
}

func (*SimpleError) Tag() Tag {
	return SimpleErrorTag
}

func (e *SimpleError) Marshal(w io.Writer) error {
	return writeSimple(e.Tag(), []byte(e.errorStr), w)
}

func (e *SimpleError) Equal(v Value) bool {
	if v, ok := v.(*SimpleError); ok {
		return e == v || e.errorStr == v.errorStr
	}
	return false
}

func readBulkError(r *Reader) (*BulkError, error) {
	data, err := r.readBulk()
	if err != nil {
		return nil, err
	}
	return MakeBulkError(string(data)), nil
}

func MakeBulkError(s string) *BulkError {
	return &BulkError{errorStr(s)}
}

func (*BulkError) Tag() Tag {
	return BulkErrorTag
}

func (e *BulkError) Marshal(w io.Writer) error {
	return writeBulk(e.Tag(), []byte(e.errorStr), w)
}

func (e *BulkError) Equal(v Value) bool {
	if v, ok := v.(*BulkError); ok {
		return e == v || e.errorStr == v.errorStr
	}
	return false
}

func (e errorStr) String() string {
	_, m := e.split()
	return m
}

func (e errorStr) Prefix() string {
	p, _ := e.split()
	return p
}

func (e errorStr) Error() string {
	return string(e)
}

func (e errorStr) split() (string, string) {
	matches := prefixedError.FindStringSubmatch(string(e))
	if len(matches) > 0 {
		return matches[1], matches[2]
	}
	return "", string(e)
}
