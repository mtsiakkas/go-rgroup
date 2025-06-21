package response

import "fmt"

type HandlerResponse struct {
	Data       any
	HttpStatus int
	LogMessage string
}

func (r *HandlerResponse) WithHttpStatus(code int) *HandlerResponse {
	r.HttpStatus = code
	return r
}

func (r *HandlerResponse) WithMessage(message string, args ...any) *HandlerResponse {
	r.LogMessage = fmt.Sprintf(message, args...)
	return r
}
