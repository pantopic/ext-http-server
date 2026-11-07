package wazero_http_server

import (
	"context"
	"io"
	"net/http"

	"github.com/tetratelabs/wazero/api"

	"github.com/pantopic/wazero-pool"
)

type httpHandler struct {
	ctx  context.Context
	pool wazeropool.Instance
	init []ContextCopier
}

func NewHttpHandler(ctx context.Context, pool wazeropool.Instance, init []ContextCopier) *httpHandler {
	return &httpHandler{
		ctx:  ctx,
		pool: pool,
		init: init,
	}
}

func (h *httpHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	meta := get[*meta](h.ctx, ctxKeyMeta)
	done := make(chan struct{})
	var called bool
	body, _ := io.ReadAll(r.Body)
	ctx := context.WithValue(h.ctx, ctxKeyReq, r)
	ctx = context.WithValue(ctx, ctxKeyResp, w)
	ctx = context.WithValue(ctx, ctxKeyBody, body)
	ctx = context.WithValue(ctx, ctxKeyDone, func() {
		if !called {
			close(done)
			called = true
		} else {
			println(`called`)
		}
	})
	for _, cc := range h.init {
		ctx = cc(ctx, h.ctx)
	}
	h.pool.Run(func(mod api.Module) {
		fn := mod.ExportedFunction(`__http_server_handle`)
		if fn == nil {
			w.WriteHeader(http.StatusNotImplemented)
			return
		}
		setHeader(mod, meta, []byte(r.Method+" "+r.URL.Path))
		writeUint32(mod, meta.ptrBodyLen, uint32(len(body)))
		setResCode(mod, meta, 200)
		fn.Call(ctx)
	})
	<-done
}
