package git

// VerifyResponse represents a response of SCM auth verify
type VerifyResponse struct {
	// Message is the detail of the result
	Message string `json:"message"`
	// Code represents a group of cases
	Code int `json:"code"`
	// Errors contain all related errors
	Errors []interface{} `json:"errors"`
}

func VerifyPass() *VerifyResponse {
	return &VerifyResponse{
		Message: "ok",
	}
}

func VerifyFailed(message string, code int, err error) *VerifyResponse {
	return &VerifyResponse{
		Message: message,
		Code:    code,
		Errors:  []interface{}{err},
	}
}

func VerifyResult(message string, code int, err error) *VerifyResponse {
	if err == nil {
		return VerifyPass()
	}
	return VerifyFailed(message, code, err)
}
