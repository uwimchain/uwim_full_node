package api_error

import (
	"encoding/json"
	"errors"
)

type ApiError struct {
	Code int    `json:"code"`
	Text string `json:"text"`
}

func NewApiError(code int, text string) *ApiError {
	return &ApiError{Code: code, Text: text}
}

func (apiError *ApiError) AddError() error {
	jsonString, _ := json.Marshal(apiError)
	return errors.New(string(jsonString))
}
