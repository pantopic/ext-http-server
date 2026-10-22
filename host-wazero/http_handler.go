package wazero_http_server

import (
	"context"
	"io"
	"net/http"

	"github.com/tetratelabs/wazero/api"
	"google.golang.org/grpc"

	"github.com/pantopic/wazero-pool"
)

type httpHandler struct {
	ctx        context.Context
	grpcServer *grpc.Server
	pool       wazeropool.Instance
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
		setBody(mod, meta, body)
		setResCode(mod, meta, 200)
		fn.Call(h.ctx)
		if code := int(getResCode(mod, meta)); code != 0 {
			w.WriteHeader(code)
		}
		w.Write(getBodyCopy(mod, meta))
	})
}
