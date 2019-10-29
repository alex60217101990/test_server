package main

import (
	"context"
	"flag"

	"github.com/alex60217101990/test_server/internal/server"
)

var (
	addr = flag.String("addr", "0.0.0.0:7755", "HTTP Server address")
)

func main() {
	flag.Parse()

	ctx, closeCtx := context.WithCancel(context.Background())

	httpServ := server.NewServer(*addr, server.SetContext(ctx))
	httpServ.Run()

	defer func() {
		closeCtx()
		httpServ.Close()
	}()
}
