package errors

type InternalError struct {
	Err error
}

func (i InternalError) Error() string {
	return i.Err.Error()
}

func (i InternalError) Unwrap() error {
	return i.Err
}

func NewInternal(err error) error {
	return InternalError{
		Err: err,
	}
}
