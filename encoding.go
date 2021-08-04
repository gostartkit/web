package web

// QueryMarshaler is the interface implemented by an object that can
// marshal itself into a query form.
//
// MarshalQuery encodes the receiver into a query form and returns the result.
type QueryMarshaler interface {
	MarshalQuery() (data []byte, err error)
}

// QueryUnmarshaler is the interface implemented by an object that can
// unmarshal a query representation of itself.
//
// UnmarshalQuery must be able to decode the form generated by MarshalQuery.
// UnmarshalQuery must copy the data if it wishes to retain the data
// after returning.
type QueryUnmarshaler interface {
	UnmarshalQuery(data []byte) error
}

// FormMarshaler is the interface implemented by an object that can
// marshal itself into a form form.
//
// MarshalForm encodes the receiver into a form form and returns the result.
type FormMarshaler interface {
	MarshalForm() (data []byte, err error)
}

// FormUnmarshaler is the interface implemented by an object that can
// unmarshal a form representation of itself.
//
// UnmarshalForm must be able to decode the form generated by MarshalForm.
// UnmarshalForm must copy the data if it wishes to retain the data
// after returning.
type FormUnmarshaler interface {
	UnmarshalForm(data []byte) error
}