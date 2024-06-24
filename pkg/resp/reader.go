package resp

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"maps"
	"strconv"
)

type (
	Reader struct {
		input        *bufio.Reader
		readers      map[Tag]ReaderFunc
		nesting      int
		v2Compatible bool
	}

	ReaderOption func(*Reader)

	ReaderFunc func(*Reader) (Value, error)
)

const (
	CR byte = 13
	LF byte = 10
)

// Error messages
const (
	ErrEmptyInput        = "ERR empty input: %e"
	ErrUnknownTag        = "ERR unknown tag: %s"
	ErrInvalidNesting    = "ERR invalid nesting: %s"
	ErrInvalidLength     = "ERR invalid length: %d"
	ErrInvalidTerminator = "ERR invalid terminator: %v"
)

var (
	v2Null    = []byte{'-', '1', CR, LF}
	v2NullLen = len(v2Null)

	defaultReaderOptions = []ReaderOption{
		DefaultReaders,
	}
)

// NewReader configures a new RESP Reader
func NewReader(r *bufio.Reader, opts ...ReaderOption) *Reader {
	res := &Reader{
		input:   r,
		readers: map[Tag]ReaderFunc{},
	}
	for _, opt := range append(defaultReaderOptions, opts...) {
		opt(res)
	}
	return res
}

// DefaultReaders adds the default RESP readers to the Reader
var DefaultReaders = WithReaderFuncs(map[Tag]ReaderFunc{
	SimpleStringTag:   asValue(readSimpleString),
	SimpleErrorTag:    asValue(readSimpleError),
	IntegerTag:        asValue(readInteger),
	BulkStringTag:     asValue(readBulkString),
	ArrayTag:          asValue(readArray),
	NullTag:           asValue(readNull),
	BooleanTag:        asValue(readBoolean),
	DoubleTag:         asValue(readDouble),
	BigNumberTag:      asValue(readBigNumber),
	BulkErrorTag:      asValue(readBulkError),
	VerbatimStringTag: asValue(readVerbatimString),
	MapTag:            asValue(readMap),
	AttributeTag:      asValue(readAttribute),
	SetTag:            asValue(readSet),
	PushTag:           asValue(readPush),
})

// V2Compatible enables V2 compatibility mode
func V2Compatible(r *Reader) {
	r.v2Compatible = true
}

func WithReaderFuncs(m map[Tag]ReaderFunc) ReaderOption {
	readers := maps.Clone(m)
	return func(r *Reader) {
		for k, v := range readers {
			r.readers[k] = v
		}
	}
}

// Next returns the next parsed value from the RESP ReaderFunc, or an error
func (r *Reader) Next() (Value, error) {
	t, err := r.input.ReadByte()
	if err != nil {
		return nil, fmt.Errorf(ErrEmptyInput, err)
	}
	tag := Tag(t)
	if r.isV2Null(tag) {
		return NullValue, nil
	}
	if fn, ok := r.readers[tag]; ok {
		wasNested := r.nesting > 0
		r.nesting++
		res, err := fn(r)
		r.nesting--
		if err == nil && wasNested {
			if _, ok := res.(TopLevelOnly); ok {
				return nil, fmt.Errorf(ErrInvalidNesting, tag)
			}
		}
		if tag == AttributeTag {
			return r.Next()
		}
		return res, err
	}
	return nil, fmt.Errorf(ErrUnknownTag, tag)
}

func (r *Reader) readSimple() ([]byte, error) {
	var buf bytes.Buffer
	for {
		data, err := r.input.ReadBytes(LF)
		if err != nil {
			return nil, err
		}
		ld := len(data) - 2
		if ld >= 0 && data[ld] == CR {
			pfx := data[:ld]
			if buf.Len() == 0 {
				return pfx, nil
			}
			return append(buf.Bytes(), pfx...), nil
		}
		buf.Write(data)
	}
}

func (r *Reader) readBulk() ([]byte, error) {
	l, err := r.readLen()
	if err != nil {
		if l == -1 {
			return nil, nil
		}
		return nil, err
	}
	data := make([]byte, l)
	_, err = io.ReadFull(r.input, data)
	if err != nil {
		return nil, err
	}
	if err = r.readNewline(); err != nil {
		return nil, err
	}
	return data, nil
}

func (r *Reader) readValues() (Values, error) {
	s, err := r.readLen()
	if err != nil {
		return nil, err
	}
	res := make(Values, s)
	for i := 0; i < s; i++ {
		val, err := r.Next()
		if err != nil {
			return nil, err
		}
		res[i] = val
	}
	return res, nil
}

func (r *Reader) readLen() (int, error) {
	i, err := r.readInt64()
	if err != nil {
		return 0, err
	}
	res := int(i)
	if int64(res) != i || res < 0 {
		return 0, fmt.Errorf(ErrInvalidLength, i)
	}
	return res, nil
}

func (r *Reader) readInt64() (int64, error) {
	data, err := r.readSimple()
	if err != nil {
		return 0, err
	}
	i, err := strconv.ParseInt(string(data), 10, 64)
	if err != nil {
		return 0, err
	}
	return i, nil
}

func (r *Reader) readNewline() error {
	data := make([]byte, 2)
	if _, err := io.ReadFull(r.input, data); err != nil {
		return err
	}
	if data[0] != CR || data[1] != LF {
		return fmt.Errorf(ErrInvalidTerminator, data)
	}
	return nil
}

func (r *Reader) isV2Null(tag Tag) bool {
	if !r.v2Compatible || !v2Nullable(tag) {
		return false
	}
	data, err := r.input.Peek(v2NullLen)
	if err == nil && bytes.Equal(data, v2Null) {
		_, _ = r.input.Discard(v2NullLen)
		return true
	}
	return false
}

func v2Nullable(t Tag) bool {
	return t == ArrayTag || t == BulkStringTag
}

func asValue[T Value](f func(*Reader) (T, error)) ReaderFunc {
	return func(r *Reader) (Value, error) {
		return f(r)
	}
}
