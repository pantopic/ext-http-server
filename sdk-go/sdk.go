package http_server

func Init(opts ...Option) {
	for _, opt := range opts {
		opt()
	}
	header = make([]byte, int(headerCap))
	body = make([]byte, int(bodyCap))
}
