package gemini

import (
	"net/http"
)

// Request 接口
type Request interface {
	ToHttpRequest(string) (*http.Request, error)
	GetEmptyResponse() Response
}

// Response 接口
type Response interface {
	FromHttpRequest(*http.Response) error
}
