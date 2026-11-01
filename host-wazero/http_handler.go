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
}

func NewHttpHandler(ctx context.Context, pool wazeropool.Instance) *httpHandler {
	return &httpHandler{
		ctx:  ctx,
		pool: pool,
	}
}

func (h *httpHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.pool.Run(func(mod api.Module) {
		fn := mod.ExportedFunction(`__http_server_handle`)
		if fn == nil {
			w.WriteHeader(http.StatusNotImplemented)
			return
		}
		meta := get[*meta](h.ctx, ctxKeyMeta)
		body, _ := io.ReadAll(r.Body)
		setHeader(mod, meta, []byte(r.Method+" "+r.URL.Path))
		writeUint32(mod, meta.ptrBodyLen, uint32(len(body)))
		setResCode(mod, meta, 200)
		ctx := context.WithValue(h.ctx, ctxKeyReq, r)
		ctx = context.WithValue(ctx, ctxKeyResp, w)
		ctx = context.WithValue(ctx, ctxKeyBody, body)
		fn.Call(ctx)
	})
}
