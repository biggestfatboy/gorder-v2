package consts

const (
	ErrnoSuccess              = 0
	ErrnoUnknownError         = 1
	ErrnoBindRequestError     = 1000
	ErrnoRequestValidateError = 1001
)

var ErrMSg = map[int]string{
	ErrnoSuccess:              "success",
	ErrnoUnknownError:         "unknown error",
	ErrnoBindRequestError:     "binding request error",
	ErrnoRequestValidateError: "request validate error",
}
