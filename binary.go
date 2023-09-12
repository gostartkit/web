package web

// binaryReader decode data from binary
func binaryReader(c *WebContext, v Data) error {
	if app().binaryReader != nil {
		return app().binaryReader(c, v)
	}
	return ErrBinaryReaderNotImplemented
}

// binaryWriter encode data to binary
func binaryWriter(c *WebContext, v Data) error {
	if app().binaryWriter != nil {
		return app().binaryWriter(c, v)
	}
	return ErrBinaryWriterNotImplemented
}
