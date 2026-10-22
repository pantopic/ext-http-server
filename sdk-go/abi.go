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

type handlerFunc func(header, path, body []byte) (code int, res []byte)

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

func getBody() []byte {
	return body[:bodyLen]
}

func getHeader() []byte {
	return header[:headerLen]
}

func getReqHeader() (header, url []byte) {
	parts := bytes.Split(getHeader(), []byte{' '})
	if len(parts) != 2 {
		resCode = 400
		setBody([]byte(`Invalid header: ` + string(getHeader())))
		return
	}
	return parts[0], parts[1]
}

//export __http_server_handle
func __http_server_handle() {
	if handler == nil {
		resCode = 405
		return
	}
	method, path := getReqHeader()
	code, b := handler(method, path, getBody())
	setBody(b)
	resCode = uint32(code)
}

//go:wasm-module pantopic/ext-http-server
//export __http_server_send
func send(uint64)

// Fix for lint rule `unusedfunc`
var _ = __http_server
var _ = __http_server_handle
var _ = send
