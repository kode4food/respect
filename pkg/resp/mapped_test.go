package resp_test

import (
	"fmt"
	"testing"

	"github.com/kode4food/respect/pkg/resp"
	"github.com/stretchr/testify/assert"
)

func TestMap(t *testing.T) {
	as := assert.New(t)

	testCases := []struct {
		input    string
		expected *resp.Map
		contains []string
	}{
		{"%0\r\n", resp.EmptyMap, []string{}},
		{
			"%1\r\n+only\r\n:1\r\n",
			resp.MakeMap(map[resp.SimpleString]resp.Value{
				"only": resp.Integer(1),
			}),
			[]string{"+only\r\n:1\r\n"},
		},
		{
			"%2\r\n+first\r\n:1\r\n+second\r\n:2\r\n",
			resp.MakeMap(map[resp.SimpleString]resp.Value{
				"first":  resp.Integer(1),
				"second": resp.Integer(2),
			}),
			[]string{"+first\r\n:1\r\n", "+second\r\n:2\r\n"},
		},
	}

	for _, tc := range testCases {
		v, err := readFromString(tc.input)
		as.Nil(err)
		as.Equal(resp.MapTag, v.Tag())
		as.True(tc.expected.Equal(v.(*resp.Map)))
		str := marshalToString(v)
		pfx := fmt.Sprintf("%%%d\r\n", len(tc.contains))
		as.Equal(pfx, str[0:len(pfx)])
		for _, c := range tc.contains {
			as.Contains(str, c)
		}
	}
}

func TestMapGet(t *testing.T) {
	as := assert.New(t)

	v, err := readFromString("%2\r\n+first\r\n:1\r\n+second\r\n:2\r\n")
	as.Nil(err)
	m := v.(*resp.Map)

	v, ok := m.Get(resp.SimpleString("first"))
	as.True(ok)
	as.True(resp.Integer(1).Equal(v))

	v, ok = m.Get(resp.SimpleString("second"))
	as.True(ok)
	as.True(resp.Integer(2).Equal(v))

	_, ok = m.Get(resp.SimpleString("third"))
	as.False(ok)
}

func testMapGet(t *testing.T, m *resp.Map, key, expected resp.Value) {
	as := assert.New(t)
	v, ok := m.Get(key)
	if expected == nil {
		as.False(ok)
		as.Nil(v)
		return
	}
	as.True(ok)
	as.True(expected.Equal(v))
}

func TestMapAggregates(t *testing.T) {
	as := assert.New(t)

	m := resp.MakeMapFromPairs(
		[2]resp.Value{resp.SimpleString("first"), resp.Integer(1)},
		[2]resp.Value{resp.SimpleString("second"), resp.Integer(2)},
		[2]resp.Value{
			resp.MakeArray(resp.Integer(1), resp.Integer(2)),
			resp.SimpleString("1 and 2"),
		},
		[2]resp.Value{
			resp.MakeMap(map[resp.SimpleString]resp.Value{
				"one": resp.Integer(1),
				"two": resp.Integer(2),
			}),
			resp.SimpleString("one:1 two:2"),
		},
		[2]resp.Value{
			resp.MakeMap(map[resp.SimpleString]resp.Value{
				"three": resp.Integer(3),
				"four":  resp.Integer(4),
			}),
			resp.SimpleString("three:3 four:4"),
		},
		[2]resp.Value{
			resp.MakeSet(resp.Integer(1), resp.Integer(2)),
			resp.SimpleString("data of 1 and 2"),
		},
	)

	as.Equal(6, m.Count())
	testMapGet(t, m, resp.SimpleString("first"), resp.Integer(1))
	testMapGet(t, m, resp.SimpleString("second"), resp.Integer(2))
	testMapGet(t, m,
		resp.MakeArray(resp.Integer(1), resp.Integer(2)),
		resp.SimpleString("1 and 2"),
	)
	testMapGet(t, m,
		resp.MakeMap(map[resp.SimpleString]resp.Value{
			"one": resp.Integer(1),
			"two": resp.Integer(2),
		}),
		resp.SimpleString("one:1 two:2"),
	)
	testMapGet(t, m,
		resp.MakeMap(map[resp.SimpleString]resp.Value{
			"three": resp.Integer(3),
			"four":  resp.Integer(4),
		}),
		resp.SimpleString("three:3 four:4"),
	)
	testMapGet(t, m,
		resp.MakeSet(resp.Integer(1), resp.Integer(2)),
		resp.SimpleString("data of 1 and 2"),
	)
}
