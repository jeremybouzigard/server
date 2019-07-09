package server

// ErrorResponse provides information about problems encountered while
// performing an operation.
type ErrorResponse struct {
	Errors []Error `json:"errors,omitempty"`
}

// NewInternalServerError creates an error with 500 HTTP status code.
func NewInternalServerError() *Error {
	e := &Error{
		Status: "500",
		Title:  "Internal Server Error",
		Detail: "There is an error processing the request."}
	return e
}

// NewStatusNotFoundError creates an error with 404 HTTP status code.
func NewStatusNotFoundError() *Error {
	e := &Error{
		Status: "404",
		Title:  "Not Found",
		Detail: "The requested resource does not exist."}
	return e
}
