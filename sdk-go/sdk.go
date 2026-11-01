package http_server

import (
	"bytes"
	"strings"
	"unsafe"
)

func Init(opts ...Option) {
	for _, opt := range opts {
		opt()
	}
	header = make([]byte, int(headerCap))
}

func GetHeaders(names ...string) map[string]string {
	headers := []byte(strings.Join(names, "\n"))
	_get_headers(uint32(uintptr(unsafe.Pointer(&headers[0]))), uint32(len(headers)))
	res := make(map[string]string)
	for line := range bytes.Lines(headerBytes()) {
		parts := bytes.SplitN(line, []byte{':'}, 2)
		if len(parts) != 2 {
			panic("invalid header line: " + string(line))
		}
		name := strings.TrimSpace(string(parts[0]))
		value := strings.TrimSpace(string(parts[1]))
		res[name] = value
	}
	return res
}

func Send(code int, body []byte) {
	resCode = uint32(code)
	_send(uint32(uintptr(unsafe.Pointer(&body[0]))), uint32(len(body)))
}

func BodyRead(buf []byte) []byte {
	if buf == nil {
		buf = make([]byte, int(bodyLen))
	}
	_read_body(uint32(uintptr(unsafe.Pointer(&buf[0]))), uint32(len(buf)))
	return buf[:int(bodyLen)]
}

func BodyLen() int {
	return int(bodyLen)
}
