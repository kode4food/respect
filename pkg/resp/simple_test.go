package resp_test

import (
	"testing"

	"github.com/kode4food/respect/pkg/resp"
	"github.com/stretchr/testify/assert"
)

func TestBoolean(t *testing.T) {
	as := assert.New(t)

	testCases := []struct {
		input    string
		expected bool
	}{
		{"#t\r\n", true},
		{"#f\r\n", false},
	}

	for _, tc := range testCases {
		v, err := resp.ReadString(tc.input)
		as.Nil(err)
		as.Equal(resp.BooleanTag, v.Tag())
		as.Equal(tc.expected, v.(resp.Boolean).Bool())
		as.Equal(tc.input, resp.ToString(v))
	}
}

func TestBooleanErrors(t *testing.T) {
	as := assert.New(t)

	testCases := []struct {
		input string
		err   string
	}{
		{"#x\r\n", "invalid boolean: x"},
		{"#true\r\n", "invalid length: 4"},
		{"#x", "EOF"},
		{"#", "EOF"},
	}

	for _, tc := range testCases {
		v, err := resp.ReadString(tc.input)
		as.Nil(v)
		as.NotNil(err)
		as.ErrorContains(err, tc.err)
	}
}

func TestNull(t *testing.T) {
	as := assert.New(t)

	v, err := resp.ReadString("_\r\n")
	as.Equal(resp.NullValue, v)
	as.Equal(resp.NullTag, v.Tag())
	as.Nil(err)
	as.Equal("_\r\n", resp.ToString(v))

	v, err = resp.ReadString("_blah\r\n")
	as.Equal(resp.NullValue, v)
	as.NotNil(err)
	as.ErrorContains(err, "invalid length: 4")

	v, err = resp.ReadString("_")
	as.Equal(resp.NullValue, v)
	as.NotNil(err)
	as.ErrorContains(err, "EOF")
}
