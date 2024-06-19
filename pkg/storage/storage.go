package storage

import (
	"errors"
	"fmt"
	"strings"

	"github.com/kode4food/respect/pkg/resp"
)

type (
	Storage interface {
		// Get retrieves a value from the storage
		Get(Key) (resp.Value, error)

		// Set stores a value in the storage
		Set(Key, resp.Value) (resp.Value, error)

		// Delete removes a value from the storage
		Delete(Key) (resp.Value, error)

		// Exists checks if a key exists in the storage
		Exists(Key) (bool, error)

		// IterateKeys returns an iterator over all keys in the Storage
		// starting at the provided Key acting as a prefix, inclusive
		IterateKeys(Key, Accept[Key]) error
	}

	Accept[T any] func(T) error

	Pair struct {
		Value resp.Value
		Key   Key
	}

	Key []resp.BulkString

	Iter[T any] interface {
		Next() (T, error)
		Close() error
	}
)

// Error messages
const (
	ErrEmptyKey   = "empty key"
	ErrInvalidKey = "invalid key type: %s"
)

var (
	// EmptyKey represents an empty Key. Only valid as a prefix
	EmptyKey = Key{}

	// StopIteration is returned by an Accept to signal that the iteration
	// operation should be stopped. This is not an error condition, and won't
	// be propagated outside the iteration operation
	StopIteration = errors.New("stop iteration")
)

func AsKey(k resp.Value) (Key, error) {
	switch k := k.(type) {
	case resp.BulkString:
		return Key{k}, nil
	case *resp.Array:
		arr := k.Elements()
		if len(arr) == 0 {
			return nil, fmt.Errorf(ErrEmptyKey)
		}
		res := make(Key, len(arr))
		for i, e := range arr {
			if e, ok := e.(resp.BulkString); ok {
				res[i] = e
				continue
			}
			return nil, fmt.Errorf(ErrInvalidKey, e.Tag())
		}
		return res, nil
	default:
		return nil, fmt.Errorf(ErrInvalidKey, k.Tag())
	}
}

func (k Key) String() string {
	if len(k) == 1 {
		return (string)(k[0])
	}
	var buf strings.Builder
	buf.WriteString((string)(k[0]))
	for _, e := range k[1:] {
		buf.WriteByte(0)
		buf.WriteString((string)(e))
	}
	return buf.String()
}

func (k Key) Equal(other Key) bool {
	if len(k) != len(other) {
		return false
	}
	for i, e := range k {
		if !e.Equal(other[i]) {
			return false
		}
	}
	return true
}
