package resp_test

import (
	"math/big"
	"testing"

	"github.com/kode4food/respect/pkg/resp"
	"github.com/stretchr/testify/assert"
)

func TestInteger(t *testing.T) {
	as := assert.New(t)

	testCases := []struct {
		input    string
		expected resp.Integer
		output   string
	}{
		{":0\r\n", resp.Integer(0), ":0\r\n"},
		{":1000\r\n", resp.Integer(1000), ":1000\r\n"},
		{":+1000\r\n", resp.Integer(1000), ":1000\r\n"},
		{":-37\r\n", resp.Integer(-37), ":-37\r\n"},
	}

	for _, tc := range testCases {
		v, err := resp.ReadString(tc.input)
		as.Nil(err)
		as.Equal(resp.IntegerTag, v.Tag())
		as.True(tc.expected.Equal(v))
		as.Equal(tc.output, resp.ToString(v))
	}
}

func TestDouble(t *testing.T) {
	as := assert.New(t)

	testCases := []struct {
		input    string
		expected resp.Double
		output   string
	}{
		{",0\r\n", resp.Double(0), ",0\r\n"},
		{",0.0\r\n", resp.Double(0), ",0\r\n"},
		{",1000.12\r\n", resp.Double(1000.12), ",1000.12\r\n"},
		{",+1000.12\r\n", resp.Double(1000.12), ",1000.12\r\n"},
		{",-37.59\r\n", resp.Double(-37.59), ",-37.59\r\n"},
		{",3.14159\r\n", resp.Double(3.14159), ",3.14159\r\n"},
	}

	for _, tc := range testCases {
		v, err := resp.ReadString(tc.input)
		as.Nil(err)
		as.Equal(resp.DoubleTag, v.Tag())
		as.True(tc.expected.Equal(v))
		as.Equal(tc.output, resp.ToString(v))
	}
}

func newBigInt(i any) *resp.BigNumber {
	switch i := i.(type) {
	case int:
		return (*resp.BigNumber)(new(big.Int).SetInt64(int64(i)))
	case string:
		if res, ok := new(big.Int).SetString(i, 10); ok {
			return (*resp.BigNumber)(res)
		}
		panic("invalid big number")
	default:
		panic("unsupported type")
	}
}

func TestBigNumber(t *testing.T) {
	as := assert.New(t)

	testCases := []struct {
		input    string
		expected *resp.BigNumber
		output   string
	}{
		{"(0\r\n", resp.ZeroBigNumber, "(0\r\n"},
		{"(1000\r\n", newBigInt(1000), "(1000\r\n"},
		{"(+1000\r\n", newBigInt(1000), "(1000\r\n"},
		{"(-37\r\n", newBigInt(-37), "(-37\r\n"},
		{"(1234567890123456789012345678901234567890\r\n", newBigInt("1234567890123456789012345678901234567890"), "(1234567890123456789012345678901234567890\r\n"},
	}

	for _, tc := range testCases {
		v, err := resp.ReadString(tc.input)
		as.Nil(err)
		as.Equal(resp.BigNumberTag, v.Tag())
		as.True(tc.expected.Equal(v))
		as.Equal(tc.output, resp.ToString(v))
	}
}
