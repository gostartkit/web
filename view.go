package web

// viewWriter encode data to html
func viewWriter(ctx *Context, v interface{}) error {
	if app().viewWriter != nil {
		return app().viewWriter(ctx, v)
	}
	return ErrViewWriterNotImplemented
}
