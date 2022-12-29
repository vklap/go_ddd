package ddd

const StatusCodeNotFound = "not_found"
const StatusCodeBadRequest = "bad_request"

type Error struct {
	message    string
	statusCode string
}

func (e *Error) Error() string {
	return e.message
}

func (e *Error) StatusCode() string {
	return e.statusCode
}

func NewError(message string, statusCode string) *Error {
	return &Error{message: message, statusCode: statusCode}
}
