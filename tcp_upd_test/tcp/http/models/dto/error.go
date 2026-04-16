package dto

type ValidationErrorResponse struct {
	Error struct {
		Code    string       `json:"code"`
		Message string       `json:"message"`
		Fields  []FieldError `json:"fields"`
	} `json:"error"`
}

type FieldError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

type ServerError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}
