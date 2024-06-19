package command

import (
	"fmt"
	"strings"

	"github.com/kode4food/respect/pkg/resp"
)

type (
	// Handler is a function that processes a command
	Handler func(Responder, ...resp.Value) error

	// Handlers is a map of command verbs to their respective Handler
	Handlers map[resp.BulkString]Handler

	// handlers is the internal representation of Handler mapping, as handlers
	// are case-insensitive, and the Handlers type is essentially a DTO
	handlers Handlers
)

// Error messages
const (
	ErrExpectedArray      = "WRONGTYPE expected array"
	ErrEmptyCommand       = "WRONGTYPE empty command"
	ErrExpectedBulkString = "WRONGTYPE expected bulk string as command"
	ErrUnknownCommand     = "ERR unknown command '%s'"
	ErrCommandProcessing  = "ERR error processing %s. %w"
)

// NewHandler creates a new Handler from a Handlers map
func NewHandler(h Handlers) Handler {
	return Wrap(h, NoHandler)
}

// NoHandler raises the unknown command error, and can be used to terminate a
// command chain
func NoHandler(_ Responder, args ...resp.Value) error {
	return resp.MakeError(fmt.Sprintf(ErrUnknownCommand, args[0]))
}

// Wrap creates a new Handler from a Handlers map, falling back to a wrapped
// Handler if none are found
func Wrap(h Handlers, wrapped Handler) Handler {
	i := h.toInternal()
	return func(c Responder, args ...resp.Value) error {
		if len(args) == 0 {
			return resp.MakeError(ErrEmptyCommand)
		}
		if args[0].Tag() != resp.BulkStringTag {
			return resp.MakeError(ErrExpectedBulkString)
		}
		verb := normalizeVerb(args[0].(resp.BulkString))
		if cmd, ok := i[verb]; ok {
			if err := cmd(c, args[1:]...); err != nil {
				return fmt.Errorf(ErrCommandProcessing, verb, err)
			}
			return nil
		}
		return wrapped(c, args...)
	}
}

func normalizeVerb(k resp.BulkString) resp.BulkString {
	return resp.BulkString(strings.ToUpper(string(k)))
}

func (h Handlers) toInternal() handlers {
	lh := make(handlers, len(h))
	for verb, v := range h {
		k := normalizeVerb(verb)
		lh[k] = v
	}
	return lh
}
