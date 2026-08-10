package gatewayruntime

import (
	"net/http"
)

// HTTPExchange is the small transport surface needed by an in-process
// runtime. It deliberately exposes no Gin or product-asset types.
type HTTPExchange interface {
	Request() *http.Request
	Header() http.Header
	WriteHeader(status int)
	Write(body []byte) (int, error)
	Flush()
	Written() bool
	Size() int
	SetState(key string, value any)
	State(key string) (any, bool)
}
