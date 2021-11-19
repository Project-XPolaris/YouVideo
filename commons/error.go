package commons

const (
	CodeErrorCodeNoLibrary = "2001"
	CodeValidatorError     = "3001"
	CodeParseError         = "3002"
)

type APIError struct {
	Err  error
	Code string
	Desc string
}

func (e *APIError) Error() string {
	return e.Err.Error()
}
