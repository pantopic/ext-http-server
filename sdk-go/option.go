package http_server

type Option func()

func WithHandler(h handlerFunc) Option {
	return func() {
		handler = h
	}
}
