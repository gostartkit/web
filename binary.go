package web

// binaryReader decode data from binary
func binaryReader(ctx *Context, v Data) error {
	if app().binaryReader != nil {
		return app().binaryReader(ctx, v)
	}
	return ErrBinaryReaderNotImplemented
}

// binaryWriter encode data to binary
func binaryWriter(ctx *Context, v Data) error {
	if app().binaryWriter != nil {
		return app().binaryWriter(ctx, v)
	}
	return ErrBinaryWriterNotImplemented
}
