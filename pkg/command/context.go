package command

import (
	"fmt"

	"github.com/kode4food/respect/pkg/resp"
)

// Error messages
const (
	ErrContextClosed = "ERR context closed"
)

type (
	Closer interface {
		Closed() <-chan struct{}
	}

	Requester interface {
		Closer
		Accept() <-chan resp.Value
	}

	Responder interface {
		Closer
		Emit() chan<- resp.Value
	}

	Context interface {
		Requester
		Responder
		Close() error
	}
)

// HandleNext processes the next command in the Context
func HandleNext(c Context, cmd Handler) error {
	v, err := Accept(c)
	if err != nil {
		return nil
	}
	if v.Tag() != resp.ArrayTag {
		return Emit(c, resp.MakeError(ErrExpectedArray))
	}
	arr := v.(resp.Collection).Elements()
	return cmd(c, arr...)
}

// Accept retrieves the next value from the Requester
func Accept(r Requester) (resp.Value, error) {
	select {
	case <-r.Closed():
		return nil, fmt.Errorf(ErrContextClosed)

	case value := <-r.Accept():
		return value, nil
	}
}

// Emit sends a value to the Responder
func Emit(r Responder, v resp.Value) error {
	select {
	case <-r.Closed():
		return fmt.Errorf(ErrContextClosed)

	case r.Emit() <- v:
		return nil
	}
}

func IsClosed(c Closer) bool {
	select {
	case <-c.Closed():
		return true

	default:
		return false
	}
}
