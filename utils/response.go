package utils

type BaseResponse struct {
	Success bool           `json:"success"`
	Message string         `json:"message"`
	Data    any            `json:"data,omitempty"`
	Error   *ErrorResponse `json:"error,omitempty"`
}

type ErrorResponse struct {
	Code    int    `json:"code"`
	Details string `json:"details"`
}
