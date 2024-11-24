package errors

// ResponseError 自定义响应错误
type ResponseError struct {
	Code int   // 错误码
	ERR  error // 响应错误
}

func (r *ResponseError) Error() string {
	if r.ERR != nil {
		return r.ERR.Error()
	}
	return ""
}

// NewResponseError 自定义响应错误
func NewResponseError(code int, err error) *ResponseError {
	res := &ResponseError{
		Code: code,
		ERR:  err,
	}
	return res
}
