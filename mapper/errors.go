package mapper

type RequestBindingError struct {
	Cause   error
	Message string
}

func NewRequestBindingError(cause error) *RequestBindingError {
	return &RequestBindingError{Message: "could not bind request", Cause: cause}
}

func (e *RequestBindingError) Error() string {
	return e.Message
}

func (e *RequestBindingError) Is(target error) bool {
	_, ok := target.(*RequestBindingError)
	return ok
}

func (e *RequestBindingError) Unwrap() error {
	return e.Cause
}

type ResponseBindingError struct {
	Cause   error
	Message string
}

func NewResponseBindingError(cause error) *ResponseBindingError {
	return &ResponseBindingError{Message: "could not bind response", Cause: cause}
}

func (e *ResponseBindingError) Error() string {
	return e.Message
}

func (e *ResponseBindingError) Is(target error) bool {
	_, ok := target.(*ResponseBindingError)
	return ok
}

func (e *ResponseBindingError) Unwrap() error {
	return e.Cause
}
