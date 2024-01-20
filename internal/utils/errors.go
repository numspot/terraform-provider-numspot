package utils

import (
	"encoding/json"
)

type ApiError struct {
	Errors []struct {
		Type    string `json:"Type"`
		Details string `json:"Details"`
		Code    string `json:"Code"`
	} `json:"Errors"`
	ResponseContext struct {
		RequestID string `json:"RequestId"`
	} `json:"ResponseContext"`
}

func (a ApiError) Type() string {
	return a.Errors[0].Type
}

func (a ApiError) Error() string {
	return a.Errors[0].Details
}

func HandleError(httpResponseBody []byte) error {
	apiError := ApiError{}
	err := json.Unmarshal(httpResponseBody, &apiError)
	if err != nil {
		return err
	}

	return apiError
}
