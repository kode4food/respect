package resp_test

import (
	"testing"

	"github.com/kode4food/respect/pkg/resp"
	"github.com/stretchr/testify/assert"
)

func TestVerbatimString(t *testing.T) {
	as := assert.New(t)
	input := "=15\r\ntxt:Some\nstring\r\n"
	v, err := readFromString(input)
	as.Nil(err)
	as.Equal(resp.VerbatimStringTag, v.Tag())
	as.Equal("txt", v.(*resp.VerbatimString).Encoding())
	as.Equal("Some\nstring", v.(*resp.VerbatimString).String())
	as.Equal(input, marshalToString(v))
}

func TestVerbatimStringErrors(t *testing.T) {
	as := assert.New(t)

	testCases := []struct {
		input    string
		expected resp.Value
		err      string
	}{
		{"=14\r\ntxt:hello", resp.EmptyVerbatimString, "EOF"},
		{"=3\r\ntxt\r\n", resp.EmptyVerbatimString, "invalid length: 3"},
		{"=3\r\ntxt\n\r", resp.EmptyVerbatimString, "invalid terminator: [10 13]"},
		{"=3\r\ntxt", resp.EmptyVerbatimString, "EOF"},
	}

	for _, tc := range testCases {
		v, err := readFromString(tc.input)
		as.Equal(tc.expected, v)
		if err != nil {
			as.ErrorContains(err, tc.err)
		} else {
			as.Nil(err)
		}
	}
}
