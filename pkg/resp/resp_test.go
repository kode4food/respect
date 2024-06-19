package resp_test

import (
	"bufio"
	"strings"

	"github.com/kode4food/respect/pkg/resp"
)

func readFromString(s string) (resp.Value, error) {
	p := resp.NewReader(bufio.NewReader(strings.NewReader(s)))
	return p.Next()
}

func marshalToString(v resp.Value) string {
	var sb strings.Builder
	_ = v.Marshal(&sb)
	return sb.String()
}
