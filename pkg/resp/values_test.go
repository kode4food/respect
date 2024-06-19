package resp_test

import (
	"testing"

	"github.com/kode4food/respect/pkg/resp"
	"github.com/stretchr/testify/assert"
)

func testValues(t *testing.T, v resp.Value, expected resp.Values) {
	as := assert.New(t)
	if v, ok := v.(resp.Collection); ok {
		as.Equal(expected, v.Elements())
		return
	}
	panic("not a value array")
}
