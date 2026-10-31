package http_server

import (
	"bytes"
	"strings"
)

func Init(opts ...Option) {
	for _, opt := range opts {
		opt()
	}
	header = make([]byte, int(headerCap))
	body = make([]byte, int(bodyCap))
}

func GetHeaders(names ...string) map[string]string {
	setHeader([]byte(strings.Join(names, "\n")))
	_get_headers()
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
