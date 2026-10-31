package http_server

import (
	"bytes"
	"unsafe"
)

var (
	meta      = make([]uint32, 7)
	resCode   uint32
	header    []byte
	headerCap uint32 = 4 << 10
	headerLen uint32
	body      []byte
	bodyCap   uint32 = 1 * 1024 * 1024
	bodyLen   uint32

	handler handlerFunc
)

type handlerFunc func(method, path []byte) (code int, res []byte)

//export __http_server
func __http_server() (res uint32) {
	meta[0] = uint32(uintptr(unsafe.Pointer(&headerCap)))
	meta[1] = uint32(uintptr(unsafe.Pointer(&headerLen)))
	meta[2] = uint32(uintptr(unsafe.Pointer(&header[0])))
	meta[3] = uint32(uintptr(unsafe.Pointer(&bodyCap)))
	meta[4] = uint32(uintptr(unsafe.Pointer(&bodyLen)))
	meta[5] = uint32(uintptr(unsafe.Pointer(&body[0])))
	meta[6] = uint32(uintptr(unsafe.Pointer(&resCode)))
	return uint32(uintptr(unsafe.Pointer(&meta[0])))
}

func setBody(b []byte) {
	copy(body[:len(b)], b)
	bodyLen = uint32(len(b))
}

func bodyBytes() []byte {
	return body[:bodyLen]
}

func setHeader(h []byte) {
	copy(header[:0], h)
	headerLen = uint32(len(h))
}

func headerBytes() []byte {
	return header[:headerLen]
}

func getRequestLine() (method, url []byte) {
	parts := bytes.Split(headerBytes(), []byte{' '})
	if len(parts) != 2 {
		resCode = 400
		setBody([]byte(`Invalid request line: ` + string(headerBytes())))
		return
	}
	method = parts[0]
	url = parts[1]
	return
}

//export __http_server_handle
func __http_server_handle() {
	if handler == nil {
		resCode = 405
		return
	}
	method, path := getRequestLine()
	code, b := handler(method, path)
	setBody(b)
	resCode = uint32(code)
}

//go:wasm-module pantopic/ext-http-server
//export __http_server_send
func _send(uint64)

//go:wasm-module pantopic/ext-http-server
//export __http_server_get_headers
func _get_headers()

// Fix for lint rule `unusedfunc`
var _ = __http_server
var _ = __http_server_handle
var _ = _send
var _ = _get_headers
