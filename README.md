# HTTP Server

A [wazero](https://pkg.go.dev/github.com/tetratelabs/wazero) host module, ABI and guest SDK enabling WASI modules to register as HTTP servers.

## Host Module

[![Go Reference](https://godoc.org/github.com/pantopic/ext-http-server/host?status.svg)](https://godoc.org/github.com/pantopic/ext-http-server/host)
[![Go Report Card](https://goreportcard.com/badge/github.com/pantopic/ext-http-server/host)](https://goreportcard.com/report/github.com/pantopic/ext-http-server/host)
[![Go Coverage](https://github.com/pantopic/ext-http-server/wiki/host/coverage.svg)](https://raw.githack.com/wiki/pantopic/ext-http-server/host/coverage.html)

First register the host module with the runtime

```go
import (
	"context"
	_ "embed"
	"net"
	"net/http"

	"github.com/tetratelabs/wazero"
	"github.com/tetratelabs/wazero/imports/wasi_snapshot_preview1"

	"github.com/pantopic/wazero-pool"

	"github.com/pantopic/ext-http-server/host-wazero"
)

//go:embed test\.wasm
var wasm []byte

func main() {
	ctx := context.Background()
	runtime := wazero.NewRuntimeWithConfig(ctx, wazero.NewRuntimeConfig())
	wasi_snapshot_preview1.MustInstantiate(ctx, runtime)

	hostModule := wazero_http_server.New()
	hostModule.Register(ctx, runtime)

	httpServer := http.NewServer()
	pool, err := wazeropool.New(ctx, runtime, wasm)
	if err != nil {
		panic(err)
	}
	hostModule.RegisterServices(ctx, httpServer, pool)

	httpListener, _ := net.Listen(`tcp`, `:8000`)
	httpServer.Serve(httpListener)

	// ...
}
```

## Roadmap

This project is in alpha. Breaking API changes should be expected until Beta.

- `v0.0.x` - Alpha
  - [ ] Stabilize API
- `v0.x.x` - Beta
  - [ ] Finalize API
  - [ ] Test in production
- `v1.x.x` - General Availability
  - [ ] Proven long term stability in production
