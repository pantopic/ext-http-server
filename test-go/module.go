package main

import (
	"github.com/pantopic/ext-http-server/sdk-go"
)

var (
	bodyBuf = make([]byte, 1<<10)
)

func init() {
	http_server.Init(
		http_server.WithHandler(handler),
	)
}

func main() {}

func handler(method, path []byte) {
	switch string(path) {
	case "/hello":
		http_server.Send(200, []byte("world"))
	case "/goodbye":
		http_server.Send(410, []byte("world"))
	case "/reflect":
		http_server.Send(200, http_server.BodyRead(nil))
	case "/reflect2":
		http_server.Send(200, http_server.BodyRead(bodyBuf))
	case "/reflect3":
		http_server.Send(200, http_server.BodyRead(make([]byte, http_server.BodyLen())))
	case "/request-header":
		http_server.Send(200, []byte(http_server.GetHeaders("test")["test"]))
	case "/response-header":
		http_server.Send(200, []byte("header"), http_server.WithHeaders(map[string]string{"test-key": "test-value"}))
	default:
		http_server.Send(404, []byte("not found"))
	}
}
