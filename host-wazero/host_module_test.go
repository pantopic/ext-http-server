package wazero_http_server

import (
	"bytes"
	"context"
	_ "embed"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
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
				t.Run(`404`, func(t *testing.T) {
					res, err := c.Get(fmt.Sprintf(`http://%s/nonsense`, addr))
					if err != nil {
						t.Fatalf(`%v`, err)
					}
					defer res.Body.Close()
					if res.StatusCode != 404 {
						t.Fatalf(`Incorrect status code: %d`, res.StatusCode)
					}
					b, err := io.ReadAll(res.Body)
					if err != nil {
						t.Fatalf(`%v`, err)
					}
					if string(b) != "not found" {
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
				t.Run(`body`, func(t *testing.T) {
					for _, path := range []string{"reflect", "reflect2", "reflect3"} {
						res, err := c.Post(fmt.Sprintf(`http://%s/%s`, addr, path), `text/plain`, bytes.NewReader([]byte("test-"+path)))
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
						if string(b) != "test-"+path {
							t.Fatalf(`Incorrect response body: "%s"`, string(b))
						}
					}
				})
				t.Run(`request header`, func(t *testing.T) {
					res, err := c.Do(&http.Request{
						URL: &url.URL{
							Scheme: "http",
							Host:   fmt.Sprintf(`%s`, addr),
							Path:   "/request-header",
						},
						Header: http.Header{
							"test": []string{"badonk"},
						},
					})
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
					if string(b) != "badonk" {
						t.Fatalf(`Incorrect response body: "%s"`, string(b))
					}
				})
				t.Run(`response header`, func(t *testing.T) {
					res, err := c.Get(fmt.Sprintf(`http://%s/response-header`, addr))
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
					if string(b) != "header" {
						t.Fatalf(`Incorrect response body: "%s"`, string(b))
					}
					if res.Header.Get("test-key") != "test-value" {
						t.Fatalf(`Incorrect response header: %s`, res.Header.Get("test-key"))
					}
				})
			})
		})
	}
}
