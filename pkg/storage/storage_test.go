package storage_test

import (
	"fmt"
	"math/rand/v2"
	"testing"

	"github.com/kode4food/respect/pkg/resp"
	"github.com/kode4food/respect/pkg/storage"
	"github.com/stretchr/testify/assert"
)

type (
	storageTest struct {
		*assert.Assertions
		make func() storage.Storage
		data []*testPair
		live map[string]struct{}
	}

	testPair struct {
		key   storage.Key
		value resp.Value
	}
)

func TestMemory(t *testing.T) {
	testStorage(t, storage.NewMemory)
}

func testStorage(t *testing.T, maker func() storage.Storage) {
	st := &storageTest{
		Assertions: assert.New(t),
		make:       maker,
		data:       getTestData(10000),
		live:       map[string]struct{}{},
	}
	st.standardOperations()
}

func getTestData(count int) []*testPair {
	data := make([]*testPair, count)
	for i := 0; i < count; i++ {
		data[i] = &testPair{
			key:   getRandomKey(i),
			value: getRandomValue(),
		}
		if rand.N(10) == 0 {
			data[i].value = nil
		}
	}
	return data
}

func getRandomKey(idx int) storage.Key {
	switch rand.N(2) {
	case 0:
		res, _ := storage.AsKey(
			resp.BulkString(fmt.Sprintf("key-%d-%d", rand.Uint64(), idx)),
		)
		return res
	default:
		cnt := rand.N(2) + 2
		arr := make([]resp.Value, cnt)
		for i := 0; i < cnt; i++ {
			str := fmt.Sprintf("key-%d-%d-%d", rand.Uint64(), idx, i)
			arr[i] = resp.BulkString(str)
		}
		res, _ := storage.AsKey(resp.MakeArray(arr...))
		return res
	}
}

func getRandomValue() resp.Value {
	switch rand.N(9) {
	case 0:
		return resp.SimpleString(fmt.Sprintf("str-%d", rand.Uint64()))
	case 1:
		return resp.Integer(rand.Int64())
	case 2:
		return resp.BulkString(fmt.Sprintf("bulkstr-%d", rand.Uint64()))
	case 3:
		return resp.MakeArray(getRandomValues(rand.N(10))...)
	case 4:
		return resp.NullValue
	case 5:
		return resp.Boolean(rand.N(2) == 0)
	case 6:
		return resp.Double(rand.Float64())
	case 7:
		res, _ := resp.MakeBigNumber(fmt.Sprintf("%d", rand.Uint64()))
		return res
	default:
		res, _ := resp.MakeVerbatimString(
			"txt", fmt.Sprintf("verbatim-%d", rand.Uint64()),
		)
		return res
	}
}

func getRandomValues(count int) []resp.Value {
	values := make([]resp.Value, count)
	for i := 0; i < count; i++ {
		values[i] = getRandomValue()
	}
	return values
}

func (t *storageTest) standardOperations() {
	s := t.make()
	t.notRetrieved(s)
	t.stored(s)
	t.exists(s)
	t.retrieved(s)
	t.deleted(s)
	t.retrieved(s)
	t.iterable(s)
}

func (t *storageTest) exists(s storage.Storage) {
	for _, d := range t.data {
		ok, err := s.Exists(d.key)
		if d.value == nil {
			t.ErrorContains(err, fmt.Sprintf(storage.ErrKeyNotFound, d.key))
		} else {
			t.Nil(err)
		}
		t.Equal(ok, d.value != nil)
	}
}

func (t *storageTest) notRetrieved(s storage.Storage) {
	for _, d := range t.data {
		v, err := s.Get(d.key)
		t.Nil(v)
		t.ErrorContains(err, fmt.Sprintf(storage.ErrKeyNotFound, d.key))
	}
}

func (t *storageTest) stored(s storage.Storage) {
	for _, d := range t.data {
		if d.value != nil {
			_, err := s.Set(d.key, d.value)
			t.Nil(err)
			v, err := s.Get(d.key)
			t.True(d.value.Equal(v))
			t.Nil(err)
			t.live[d.key.String()] = struct{}{}
		}
	}
}

func (t *storageTest) retrieved(s storage.Storage) {
	for _, d := range t.data {
		v, err := s.Get(d.key)
		if d.value == nil {
			t.Nil(v)
			t.ErrorContains(err, fmt.Sprintf(storage.ErrKeyNotFound, d.key))
			continue
		}
		t.NotNil(v)
		t.Nil(err)
		t.True(d.value.Equal(v))
	}
}

func (t *storageTest) deleted(s storage.Storage) {
	for _, d := range t.data {
		if d.value != nil && rand.N(10) == 0 {
			v, err := s.Delete(d.key)
			t.Nil(err)
			t.True(d.value.Equal(v))
			v, err = s.Get(d.key)
			t.Nil(v)
			t.ErrorContains(err, fmt.Sprintf(storage.ErrKeyNotFound, d.key))
			d.value = nil
			delete(t.live, d.key.String())
		}
	}
}

func (t *storageTest) iterable(s storage.Storage) {
	seen := make(map[string]struct{}, len(t.data))
	next := len(t.data)

	err := s.IterateKeys(storage.EmptyKey, func(k storage.Key) error {
		if _, ok := seen[k.String()]; ok {
			t.FailNow(fmt.Sprintf("key already seen: %s", k.String()))
		}
		seen[k.String()] = struct{}{}

		if rand.N(2) == 0 {
			orig, err := s.Get(k)
			t.Nil(err)
			t.NotNil(orig)

			nv := getRandomValue()
			prev, err := s.Set(k, nv)
			t.Equal(prev, orig)
			t.Nil(err)

			v, err := s.Get(k)
			t.Nil(err)
			t.Equal(nv, v)
		}

		if rand.N(10) == 0 {
			k := getRandomKey(next)
			v := getRandomValue()
			old, err := s.Set(k, v)
			t.Nil(old)
			t.Nil(err)

			res, err := s.Get(k)
			t.True(v.Equal(res))
			t.Nil(err)

			next++
			t.live[k.String()] = struct{}{}
		}
		return nil
	})
	t.Nil(err)
	t.Equal(len(seen), len(t.live))
}
