package http_server

type Error struct {
	code int
	msg  []byte
}

func (e Error) Error() string {
	return string(e.msg)
}

var (
	ErrMethodNotAllowed = Error{405, []byte("Method Not Allowed")}
)
