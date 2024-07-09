package resp_test

import (
	"testing"

	"github.com/kode4food/respect/pkg/resp"
	"github.com/stretchr/testify/assert"
)

func TestSimpleString(t *testing.T) {
	as := assert.New(t)

	testCases := []struct {
		input    string
		expected string
	}{
		{"+OK\r\n", "OK"},
		{"+\r\n", ""},
		{"+Embedded\nLine\r\n", "Embedded\nLine"},
	}

	for _, tc := range testCases {
		v, err := resp.ReadString(tc.input)
		as.Nil(err)
		as.Equal(resp.SimpleStringTag, v.Tag())
		as.Equal(tc.expected, v.(resp.SimpleString).String())
		as.Equal(tc.input, resp.ToString(v))
	}
}

func TestSimpleStringErrors(t *testing.T) {
	as := assert.New(t)
	v, err := resp.ReadString("+OK")
	as.Equal(resp.EmptySimpleString, v)
	as.NotNil(err)
	as.ErrorContains(err, "EOF")
}

func TestBulkString(t *testing.T) {
	as := assert.New(t)

	testCases := []struct {
		input    string
		expected string
	}{
		{"$10\r\nhello\r\nyou\r\n", "hello\r\nyou"},
		{"$0\r\n\r\n", ""},
		{"$14\r\nEmbedded\r\nLine\r\n", "Embedded\r\nLine"},
	}

	for _, tc := range testCases {
		v, err := resp.ReadString(tc.input)
		as.Nil(err)
		as.Equal(resp.BulkStringTag, v.Tag())
		as.Equal(tc.expected, v.(resp.BulkString).String())
		as.Equal(tc.input, resp.ToString(v))
	}
}

func TestBulkStringErrors(t *testing.T) {
	as := assert.New(t)

	testCases := []struct {
		input    string
		expected resp.Value
		err      string
	}{
		{"$10\r\nhello", resp.EmptyBulkString, "EOF"},
		{"$5\r\nhello\n\r", resp.EmptyBulkString, "invalid terminator: [10 13]"},
	}

	for _, tc := range testCases {
		v, err := resp.ReadString(tc.input)
		as.Equal(tc.expected, v)
		if err != nil {
			as.ErrorContains(err, tc.err)
		} else {
			as.Nil(err)
		}
	}
}
