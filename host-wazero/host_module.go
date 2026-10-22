package wazero_http_server

import (
	"context"
	"log"
	"net"
	"net/http"
	"sync"

	"github.com/soheilhy/cmux"
	"github.com/tetratelabs/wazero"
	"github.com/tetratelabs/wazero/api"

	"github.com/pantopic/wazero-pool"
)

const Name = "pantopic/ext-http-server"

var (
	ctxKeyMeta = Name + `/meta`
	ctxKeyNext = Name + `/next`
	ctxKeySend = Name + `/send`
)

type meta struct {
	ptrHeaderCap uint32
	ptrHeaderLen uint32
	ptrHeader    uint32
	ptrBodyCap   uint32
	ptrBodyLen   uint32
	ptrBody      uint32
	ptrResCode   uint32
}

type hostModule struct {
	sync.RWMutex

	module api.Module
	mux    map[net.Listener]cmux.CMux
}

type sendFunc = func(ctx context.Context, msg []byte, err error)

func New(opts ...Option) (h *hostModule) {
	h = &hostModule{
		mux: make(map[net.Listener]cmux.CMux),
	}
	for _, opt := range opts {
		opt(h)
	}
	return
}

func (h *hostModule) Name() string {
	return Name
}

// Register instantiates the host module, making it available to all module instances in this runtime
func (h *hostModule) Register(ctx context.Context, r wazero.Runtime) (err error) {
	builder := r.NewHostModuleBuilder(Name)
	register := func(name string, fn func(ctx context.Context, m api.Module, stack []uint64)) {
		builder = builder.NewFunctionBuilder().WithGoModuleFunction(api.GoModuleFunc(fn), []api.ValueType{api.ValueTypeI64}, nil).Export(name)
	}
	for name, fn := range map[string]any{
		"__http_server_send": func(ctx context.Context, msg []byte, code int) {
			get[sendFunc](ctx, ctxKeySend)(ctx, msg, err)
		},
	} {
		switch fn := fn.(type) {
		case func(context.Context, []byte, int):
			register(name, func(ctx context.Context, mod api.Module, stack []uint64) {
				meta := get[*meta](ctx, ctxKeyMeta)
				var data []byte
				if uint32(stack[0]) > 0 {
					buf, _ := mod.Memory().Read(uint32(stack[0]>>32), uint32(stack[0]))
					data = append(data[:0], buf...)
				}
				fn(ctx, data, getResCode(mod, meta))
			})
		default:
			log.Panicf("Method signature implementation missing: %#v", fn)
		}
	}
	h.module, err = builder.Instantiate(ctx)
	return
}

// InitContext retrieves the meta page from the wasm module
func (h *hostModule) InitContext(ctx context.Context, m api.Module) (context.Context, error) {
	fn := m.ExportedFunction(`__http_server`)
	if fn == nil {
		return ctx, nil
	}
	stack, err := fn.Call(ctx)
	if err != nil {
		return ctx, err
	}
	meta := &meta{}
	ptr := uint32(stack[0])
	for i, v := range []*uint32{
		&meta.ptrHeaderCap,
		&meta.ptrHeaderLen,
		&meta.ptrHeader,
		&meta.ptrBodyCap,
		&meta.ptrBodyLen,
		&meta.ptrBody,
		&meta.ptrResCode,
	} {
		*v = readUint32(m, ptr+uint32(4*i))
	}
	return context.WithValue(ctx, ctxKeyMeta, meta), nil
}

// ContextCopier populates dst context with the meta page from src context.
func (h *hostModule) ContextCopier(dst, src context.Context) context.Context {
	if v := src.Value(ctxKeyMeta); v != nil {
		dst = context.WithValue(dst, ctxKeyMeta, v.(*meta))
		if v := src.Value(ctxKeyNext); v != nil {
			dst = context.WithValue(dst, ctxKeyNext, v.(chan []byte))
		}
		if v := src.Value(ctxKeySend); v != nil {
			dst = context.WithValue(dst, ctxKeySend, v.(sendFunc))
		}
	}
	return dst
}

func (h *hostModule) ServiceType() string {
	return `http`
}

func (h *hostModule) ServerStart(ctx context.Context, lis net.Listener, tlsCrt, tlsKey string, pool wazeropool.Instance, ctxCopy ...ContextCopier) (err error) {
	m := cmux.New(lis)
	httpListener := m.Match(cmux.Any())
	httpServer := &http.Server{
		Handler: NewHttpHandler(ctx, pool),
	}
	go func() {
		if tlsCrt != "" && tlsKey != "" {
			if err = httpServer.ServeTLS(httpListener, tlsCrt, tlsKey); err != nil {
				panic(err)
			}
		} else {
			if err = httpServer.Serve(httpListener); err != nil {
				panic(err)
			}
		}
		if err = httpServer.Serve(httpListener); err != nil {
			panic(err)
		}
	}()
	go func() {
		if err := m.Serve(); err != nil {
			panic(err)
		}
	}()
	h.mux[lis] = m
	return
}

func (h *hostModule) ServerStop(ctx context.Context, lis net.Listener) (err error) {
	m, ok := h.mux[lis]
	if !ok {
		return
	}
	m.Close()
	return
}

func (h *hostModule) Stop() (err error) {
	return
}

func get[T any](ctx context.Context, key string) T {
	v := ctx.Value(key)
	if v == nil {
		log.Panicf("Context item missing %s", key)
	}
	return v.(T)
}

func getResCode(m api.Module, meta *meta) int {
	return int(readUint32(m, meta.ptrResCode))
}
func setResCode(m api.Module, meta *meta, code int) {
	writeUint32(m, meta.ptrResCode, uint32(code))
}

func headerBuf(m api.Module, meta *meta) []byte {
	return read(m, meta.ptrHeader, 0, meta.ptrHeaderCap)
}

func setHeader(m api.Module, meta *meta, header []byte) {
	copy(headerBuf(m, meta)[:len(header)], header)
	writeUint32(m, meta.ptrHeaderLen, uint32(len(header)))
}

func getBody(m api.Module, meta *meta) []byte {
	return read(m, meta.ptrBody, meta.ptrBodyLen, meta.ptrBodyCap)
}

func getBodyCopy(m api.Module, meta *meta) []byte {
	return append([]byte(nil), getBody(m, meta)...)
}

func bodyBuf(m api.Module, meta *meta) []byte {
	return read(m, meta.ptrBody, 0, meta.ptrBodyCap)
}

func setBody(m api.Module, meta *meta, msg []byte) error {
	buf := bodyBuf(m, meta)
	if len(msg) > cap(buf) {
		writeUint32(m, meta.ptrBodyLen, 0)
		return ErrRequestTooLarge
	}
	copy(buf[:len(msg)], msg)
	writeUint32(m, meta.ptrBodyLen, uint32(len(msg)))
	return nil
}

func read(m api.Module, ptr, ptrLen, ptrCap uint32) (buf []byte) {
	buf, ok := m.Memory().Read(ptr, readUint32(m, ptrCap))
	if !ok {
		log.Panicf("Memory.Read(%d, %d) out of range", ptr, ptrCap)
	}
	return buf[:readUint32(m, ptrLen)]
}

func readUint32(m api.Module, ptr uint32) (val uint32) {
	val, ok := m.Memory().ReadUint32Le(ptr)
	if !ok {
		log.Panicf("Memory.Read(%d) out of range", ptr)
	}
	return
}

func writeUint32(m api.Module, ptr uint32, val uint32) {
	if ok := m.Memory().WriteUint32Le(ptr, val); !ok {
		log.Panicf("Memory.Read(%d) out of range", ptr)
	}
}
