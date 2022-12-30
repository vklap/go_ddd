package go_ddd

// StatusCodeNotFound is a string that represents a "not found" error.
const StatusCodeNotFound = "not_found"

// StatusCodeBadRequest is a string that represents a "not found" error.
const StatusCodeBadRequest = "bad_request"

// Error is a struct that contains both a message and a status code.
type Error struct {
	message    string
	statusCode string
}

// Error returns the error message.
func (e *Error) Error() string {
	return e.message
}

// StatusCode returns the error status code.
func (e *Error) StatusCode() string {
	return e.statusCode
}

// NewError initializes a new Error instance.
func NewError(message string, statusCode string) *Error {
	return &Error{message: message, statusCode: statusCode}
}
