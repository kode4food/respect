package resp_test

import (
	"testing"

	"github.com/kode4food/respect/pkg/resp"
	"github.com/stretchr/testify/assert"
)

func TestPush(t *testing.T) {
	as := assert.New(t)

	testCases := []struct {
		input    string
		expected []resp.Value
	}{
		{">0\r\n", []resp.Value{}},
		{
			">2\r\n$5\r\nhello\r\n$5\r\nworld\r\n",
			[]resp.Value{resp.BulkString("hello"), resp.BulkString("world")},
		},
		{
			">3\r\n:1\r\n:2\r\n:3\r\n",
			[]resp.Value{resp.Integer(1), resp.Integer(2), resp.Integer(3)},
		},
	}

	for _, tc := range testCases {
		v, err := readFromString(tc.input)
		as.Nil(err)
		as.Equal(resp.PushTag, v.Tag())
		testValues(t, v, tc.expected)
		as.Equal(tc.input, marshalToString(v))
	}
}
