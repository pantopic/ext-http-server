package wazero_http_server

var (
	ErrRequestTooLarge = Error{400, "request too large"}
)

type Error struct {
	code int
	msg  string
}

func (e Error) Error() string {
	return e.msg
}
