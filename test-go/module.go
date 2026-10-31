package main

import (
	"github.com/pantopic/ext-http-server/sdk-go"
)

func init() {
	http_server.Init(
		http_server.WithHandler(handler),
	)
}

func main() {}

func handler(method, path []byte) (code int, res []byte) {
	switch string(path) {
	case "/hello":
		code = 200
		res = append(res, "world"...)
	case "/goodbye":
		code = 410
		res = append(res, "world"...)
	}
	return
}
