package jsend

const (
	StatusSuccess = "success"
	StatusFail    = "fail"
	StatusError   = "error"
)

type Response struct {
	Status  string  `json:"status" yaml:"status"`
	Data    any     `json:"data,omitempty" yaml:"data,omitempty"`
	Message *string `json:"message,omitempty" yaml:"message,omitempty"`
	Code    *string `json:"code,omitempty" yaml:"code,omitempty"`
}

func Success(data any) *Response {
	return &Response{
		Status: StatusSuccess,
		Data:   data,
	}
}

func Fail(data any) *Response {
	return &Response{
		Status: StatusFail,
		Data:   data,
	}
}

func Error(message string, data any, code string) *Response {
	return &Response{
		Status:  StatusError,
		Message: &message,
		Data:    data,
		Code:    &code,
	}
}

func SimpleErr(message string) *Response {
	return &Response{
		Status:  StatusError,
		Message: &message,
	}
}
