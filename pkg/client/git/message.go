package git

// VerifyResponse represents a response of SCM auth verify
type VerifyResponse struct {
	// Message is the detail of the result
	Message string `json:"message"`
	// Code represents a group of cases
	Code int `json:"code"`
}

func VerifyPass() *VerifyResponse {
	return &VerifyResponse{
		Message: "ok",
	}
}

func VerifyFailed(message string, code int) *VerifyResponse {
	return &VerifyResponse{
		Message: message,
		Code:    code,
	}
}

func VerifyResult(err error, code int) *VerifyResponse {
	if err == nil {
		return VerifyPass()
	}
	return VerifyFailed(err.Error(), code)
}
