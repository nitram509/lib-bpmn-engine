package apierror

type ApiError struct {
	Message string `json:"message"`
	Type    string `json:"type"`
}
