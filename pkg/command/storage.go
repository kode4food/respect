package command

import (
	"github.com/kode4food/respect/pkg/resp"
	"github.com/kode4food/respect/pkg/storage"
)

type storageOp func(storage.Storage, ...resp.Value) (resp.Value, error)

const (
	// ErrWrongArgumentCount is returned upon the wrong number of arguments
	ErrWrongArgumentCount = "wrong number of arguments: %d"
)

func Storage(s storage.Storage) Handler {
	h := storageHandlers(s)
	return NewHandler(h)
}

func StorageWrap(s storage.Storage, next Handler) Handler {
	h := storageHandlers(s)
	return Wrap(h, next)
}

func storageHandlers(s storage.Storage) Handlers {
	return Handlers{
		"GET": wrapStorageOp(s, getOp),
		"SET": wrapStorageOp(s, setOp),
		"DEL": wrapStorageOp(s, deleteOp),
	}
}

func getOp(s storage.Storage, args ...resp.Value) (resp.Value, error) {
	if len(args) != 1 {
		return nil, resp.MakeError(ErrWrongArgumentCount, 1)
	}
	key, err := storage.AsKey(args[0])
	if err != nil {
		return nil, err
	}
	return s.Get(key)
}

func setOp(s storage.Storage, args ...resp.Value) (resp.Value, error) {
	if len(args) != 2 {
		return nil, resp.MakeError(ErrWrongArgumentCount, 2)
	}
	key, err := storage.AsKey(args[0])
	if err != nil {
		return nil, err
	}
	if _, err = s.Set(key, args[1]); err != nil {
		return nil, err
	}
	return resp.OK, nil
}

func deleteOp(s storage.Storage, args ...resp.Value) (resp.Value, error) {
	if len(args) != 1 {
		return nil, resp.MakeError(ErrWrongArgumentCount, 1)
	}
	key, err := storage.AsKey(args[0])
	if err != nil {
		return nil, err
	}
	if _, err = s.Delete(key); err != nil {
		return nil, err
	}
	return resp.OK, nil
}

func wrapStorageOp(s storage.Storage, op storageOp) Handler {
	return func(r Responder, args ...resp.Value) error {
		value, err := op(s, args...)
		if err != nil {
			return resp.MakeError(err.Error())
		}
		r.Emit() <- value
		return nil
	}
}
