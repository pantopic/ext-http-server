package wazero_http_server

import (
	"context"
	_ "embed"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/tetratelabs/wazero"
	"github.com/tetratelabs/wazero/api"
	"github.com/tetratelabs/wazero/imports/wasi_snapshot_preview1"

	"github.com/pantopic/wazero-pool"
)

//go:embed test.go.wasm
var testWasmGo []byte

func TestHostModule(t *testing.T) {
	ctx := context.Background()
	r := wazero.NewRuntimeWithConfig(ctx, wazero.NewRuntimeConfig())
	wasi_snapshot_preview1.MustInstantiate(ctx, r)

	var hostModule *hostModule
	t.Run(`register`, func(t *testing.T) {
		hostModule = New()
		hostModule.Register(ctx, r)
	})

	port := 9000
	for _, tc := range []struct {
		name string
		wasm []byte
	}{
		{`testWasmGo`, testWasmGo},
		// {`testWasmZig`, testWasmZig},
	} {
		t.Run(tc.name, func(t *testing.T) {
			cfg := wazero.NewModuleConfig().WithStdout(os.Stdout)
			pool, err := wazeropool.New(ctx, r, tc.wasm,
				wazeropool.WithModuleConfig(cfg))
			if err != nil {
				t.Fatalf(`%v`, err)
			}
			pool.Run(func(mod api.Module) {
				ctx, err = hostModule.InitContext(ctx, mod)
				if err != nil {
					t.Fatalf(`%v`, err)
				}
			})
			ctxCopy := []ContextCopier{}
			port++
			addr := fmt.Sprintf(`:%d`, port)
			lis, err := net.Listen(`tcp`, addr)
			if err != nil {
				t.Fatalf(`%v`, err)
			}
			hostModule.ServerStart(ctx, lis, "", "", pool, ctxCopy...)
			c := &http.Client{Timeout: time.Second}
			t.Run(`GET`, func(t *testing.T) {
				t.Run(`200`, func(t *testing.T) {
					res, err := c.Get(fmt.Sprintf(`http://%s/hello`, addr))
					if err != nil {
						t.Fatalf(`%v`, err)
					}
					defer res.Body.Close()
					if res.StatusCode != 200 {
						t.Fatalf(`Incorrect status code: %d`, res.StatusCode)
					}
					b, err := io.ReadAll(res.Body)
					if err != nil {
						t.Fatalf(`%v`, err)
					}
					if string(b) != "world" {
						t.Fatalf(`Incorrect response body: %s`, string(b))
					}
				})
				t.Run(`410`, func(t *testing.T) {
					res, err := c.Get(fmt.Sprintf(`http://%s/goodbye`, addr))
					if err != nil {
						t.Fatalf(`%v`, err)
					}
					defer res.Body.Close()
					if res.StatusCode != 410 {
						t.Fatalf(`Incorrect status code: %d`, res.StatusCode)
					}
					b, err := io.ReadAll(res.Body)
					if err != nil {
						t.Fatalf(`%v`, err)
					}
					if string(b) != "world" {
						t.Fatalf(`Incorrect response body: %s`, string(b))
					}
				})
			})
		})
	}
}
