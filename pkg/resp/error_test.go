package resp_test

import (
	"testing"

	"github.com/kode4food/respect/pkg/resp"
	"github.com/stretchr/testify/assert"
)

func TestMakeError(t *testing.T) {
	as := assert.New(t)

	testCases := []struct {
		input    string
		expected string
		pfx, msg string
		output   resp.Value
	}{
		{
			"ERR unknown command 'asdf'",
			"-ERR unknown command 'asdf'\r\n",
			"ERR", "unknown command 'asdf'",
			resp.MakeSimpleError("ERR unknown command 'asdf'"),
		},
		{
			"SYNTAX\r\ninvalid syntax",
			"!22\r\nSYNTAX\r\ninvalid syntax\r\n",
			"SYNTAX", "invalid syntax",
			resp.MakeBulkError("SYNTAX\r\ninvalid syntax"),
		},
	}

	for _, tc := range testCases {
		v := resp.MakeError(tc.input)
		as.Equal(tc.expected, marshalToString(v))
		as.Equal(tc.pfx, v.(resp.Error).Prefix())
		as.Equal(tc.msg, v.(resp.Error).String())
		as.True(tc.output.Equal(v))
	}
}

func TestSimpleError(t *testing.T) {
	as := assert.New(t)

	testCases := []struct {
		input    string
		expected string
		pfx, msg string
	}{
		{
			"-ERR unknown command 'asdf'\r\n",
			"ERR unknown command 'asdf'",
			"ERR", "unknown command 'asdf'",
		},
		{"-\r\n", "", "", ""},
		{
			"-Embedded\nLine\r\n",
			"Embedded\nLine",
			"", "Embedded\nLine",
		},
	}

	for _, tc := range testCases {
		v, err := readFromString(tc.input)
		as.Nil(err)
		as.Equal(resp.SimpleErrorTag, v.Tag())
		as.Equal(tc.expected, v.(*resp.SimpleError).Error())
		as.Equal(tc.input, marshalToString(v))
		as.Equal(tc.pfx, v.(*resp.SimpleError).Prefix())
		as.Equal(tc.msg, v.(*resp.SimpleError).String())
	}

	as.False(
		resp.MakeSimpleError("ERR unknown command 'asdf'").Equal(
			resp.MakeBulkError("ERR unknown command 'asdf'"),
		),
	)
}

func TestBulkError(t *testing.T) {
	as := assert.New(t)

	testCases := []struct {
		input    string
		expected string
		pfx, msg string
	}{
		{
			"!22\r\nSYNTAX\r\ninvalid syntax\r\n",
			"SYNTAX\r\ninvalid syntax",
			"SYNTAX", "invalid syntax",
		},
		{"!0\r\n\r\n", "", "", ""},
		{
			"!14\r\nEmbedded\r\nLine\r\n",
			"Embedded\r\nLine",
			"", "Embedded\r\nLine",
		},
	}

	for _, tc := range testCases {
		v, err := readFromString(tc.input)
		as.Nil(err)
		as.Equal(resp.BulkErrorTag, v.Tag())
		as.Equal(tc.expected, v.(*resp.BulkError).Error())
		as.Equal(tc.input, marshalToString(v))
		as.Equal(tc.pfx, v.(*resp.BulkError).Prefix())
		as.Equal(tc.msg, v.(*resp.BulkError).String())
	}

	as.False(
		resp.MakeBulkError("ERR unknown command 'asdf'").Equal(
			resp.MakeSimpleError("ERR unknown command 'asdf'"),
		),
	)
}
