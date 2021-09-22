package api_error

import (
	"encoding/json"
	"errors"
)

type ApiError struct {
	Code int
	Text string
}

func NewApiError(code int, text string) *ApiError {
	return &ApiError{Code: code, Text: text}
}

func AddError(apiError *ApiError) error {
	jsonString, _ := json.Marshal(apiError)
	return errors.New(string(jsonString))
}
