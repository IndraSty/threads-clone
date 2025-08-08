package models

type APIResponse struct {
	Success bool        `json:"success"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
	Error   *APIError   `json:"error,omitempty"`
}

type APIError struct {
	Code    string      `json:"code"`
	Message string      `json:"message"`
	Details interface{} `json:"details,omitempty"`
}

func SuccessResponse(message string, data interface{}) APIResponse {
	return APIResponse{
		Success: true,
		Message: message,
		Data:    data,
	}
}

func ErrorResponse(code, message string, details interface{}) APIResponse {
	return APIResponse{
		Success: false,
		Message: message,
		Error: &APIError{
			Code:    code,
			Message: message,
			Details: details,
		},
	}
}
