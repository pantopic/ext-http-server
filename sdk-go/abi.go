package http_server

import (
	"bytes"
	"unsafe"
)

var (
	meta      = make([]uint32, 5)
	resCode   uint32
	header    []byte
	headerCap uint32 = 4 << 10
	headerLen uint32
	bodyLen   uint32

	handler handlerFunc
)

type handlerFunc func(method, path []byte)

//export __http_server
func __http_server() (res uint32) {
	meta[0] = uint32(uintptr(unsafe.Pointer(&headerCap)))
	meta[1] = uint32(uintptr(unsafe.Pointer(&headerLen)))
	meta[2] = uint32(uintptr(unsafe.Pointer(&header[0])))
	meta[3] = uint32(uintptr(unsafe.Pointer(&bodyLen)))
	meta[4] = uint32(uintptr(unsafe.Pointer(&resCode)))
	return uint32(uintptr(unsafe.Pointer(&meta[0])))
}

func headerBytes() []byte {
	return header[:headerLen]
}

func getRequestLine() (method, url []byte) {
	parts := bytes.Split(headerBytes(), []byte{' '})
	if len(parts) != 2 {
		resCode = 400
		res := []byte(`Invalid request line: ` + string(headerBytes()))
		_send(uint32(uintptr(unsafe.Pointer(&res[0]))), uint32(len(res)))
		return
	}
	method = parts[0]
	url = parts[1]
	return
}

//export __http_server_handle
func __http_server_handle() {
	if handler == nil {
		_sendErr(ErrMethodNotAllowed)
		return
	}
	method, path := getRequestLine()
	handler(method, path)
}

func _sendErr(err Error) {
	resCode = uint32(err.code)
	_send(uint32(uintptr(unsafe.Pointer(&err.msg[0]))), uint32(len(err.msg)))
}

//go:wasm-module pantopic/ext-http-server
//export __http_server_send
func _send(uint32, uint32)

//go:wasm-module pantopic/ext-http-server
//export __http_server_get_headers
func _get_headers(uint32, uint32)

//go:wasm-module pantopic/ext-http-server
//export __http_server_read_body
func _read_body(uint32, uint32)

// Fix for lint rule `unusedfunc`
var _ = __http_server
var _ = __http_server_handle
var _ = _send
var _ = _get_headers
var _ = _read_body
