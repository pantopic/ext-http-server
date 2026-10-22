package wazero_http_server

import (
	"context"
)

type ContextCopier = func(dst, src context.Context) context.Context
