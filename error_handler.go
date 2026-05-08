package web

// ErrorBody is the structured response payload written by JSONErrorHandler.
//
// Applications can reuse this type when they want their own ErrorHandler to keep
// the same wire format while changing logging, redaction, or status mapping.
type ErrorBody struct {
	// Code is the HTTP status code selected for the error.
	Code int `json:"code" xml:"code"`

	// Message is the error string returned to the client.
	Message string `json:"message" xml:"message"`

	// RequestID is included only when JSONErrorHandler is configured to include it
	// and RequestID middleware has populated a value.
	RequestID string `json:"request_id,omitempty" xml:"request_id,omitempty"`
}

// JSONErrorHandler returns an ErrorHandler that writes a structured JSON error body.
//
// This is opt-in and does not affect the framework's default error semantics
// unless it is installed with Application.SetErrorHandler or WithErrorHandler.
// Example:
//
//	app := web.New(web.WithErrorHandler(web.JSONErrorHandler(true)))
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
