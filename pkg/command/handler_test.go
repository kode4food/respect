package command_test

import (
	"testing"

	"github.com/kode4food/respect/pkg/command"
	"github.com/kode4food/respect/pkg/resp"
	"github.com/stretchr/testify/assert"
)

func TestNewHandler(t *testing.T) {
	as := assert.New(t)
	h := command.NewHandler(command.Handlers{
		"GET": func(r command.Responder, args ...resp.Value) error {
			return nil
		},
	})
	as.NotNil(h)
}
