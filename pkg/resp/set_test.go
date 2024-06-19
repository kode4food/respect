package resp_test

import (
	"fmt"
	"testing"

	"github.com/kode4food/respect/pkg/resp"
	"github.com/stretchr/testify/assert"
)

func TestSet(t *testing.T) {
	as := assert.New(t)

	testCases := []struct {
		input    string
		expected *resp.Set
		contains []string
	}{
		{"~0\r\n", resp.EmptySet, []string{}},
		{
			"~2\r\n$5\r\nhello\r\n$5\r\nworld\r\n",
			resp.MakeSet(resp.BulkString("hello"), resp.BulkString("world")),
			[]string{"$5\r\nhello\r\n", "$5\r\nworld\r\n"},
		},
		{
			"~3\r\n:1\r\n:2\r\n:3\r\n",
			resp.MakeSet(resp.Integer(1), resp.Integer(2), resp.Integer(3)),
			[]string{":1\r\n", ":2\r\n", ":3\r\n"},
		},
	}

	for _, tc := range testCases {
		v, err := readFromString(tc.input)
		as.Nil(err)
		as.Equal(resp.SetTag, v.Tag())
		as.True(tc.expected.Equal(v.(*resp.Set)))
		str := marshalToString(v)
		pfx := fmt.Sprintf("~%d\r\n", len(tc.contains))
		as.Equal(pfx, str[0:len(pfx)])
		for _, c := range tc.contains {
			as.Contains(str, c)
		}
	}
}
