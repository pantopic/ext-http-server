package http_server

type Option func()

func WithHandler(h handlerFunc) Option {
	return func() {
		handler = h
	}
}

func WithHeaders(m map[string]string) Option {
	return func() {
		for k, v := range m {
			header = append(header, []byte(k+": "+v+"\n")...)
		}
	}
}
