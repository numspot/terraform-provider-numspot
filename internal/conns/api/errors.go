package api

var undefinedError = "undefined error."

func (e ErrorResponse) Error() string {
	if e.Detail == nil {
		return undefinedError
	}
	return *e.Detail
}
