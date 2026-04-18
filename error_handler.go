package web

// ErrorBody is a structured API error payload for opt-in error handlers.
type ErrorBody struct {
	Code      int    `json:"code" xml:"code"`
	Message   string `json:"message" xml:"message"`
	RequestID string `json:"request_id,omitempty" xml:"request_id,omitempty"`
}

// JSONErrorHandler returns an ErrorHandler that writes a structured JSON error body.
// This is opt-in and does not affect the framework's default error semantics.
func JSONErrorHandler(includeRequestID bool) ErrorHandler {
	return func(c *Ctx, err error) error {
		body := ErrorBody{
			Code:    errCode(err),
			Message: err.Error(),
		}
		if includeRequestID {
			body.RequestID = c.RequestID()
		}

		c.SetContentType("application/json")
		c.WriteHeader(body.Code)
		return c.writeJSON(body)
	}
}
