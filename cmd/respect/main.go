package main

import (
	"github.com/kode4food/respect/pkg/command"
	"github.com/kode4food/respect/pkg/server"
	"github.com/kode4food/respect/pkg/storage"
)

func main() {
	s := storage.NewMemory()
	h := server.WithHandler(command.Storage(s))
	svr := server.NewServer(h)
	if err := svr.Start(); err != nil {
		panic(err)
	}
}
