package utils

import (
	"encoding/json"
	"errors"
	"net/http"

	"gitlab.numspot.cloud/cloud/numspot-sdk-go/pkg/numspot"
)

func ParseHTTPError(httpResponseBody []byte, statusCode int) error {
	if statusCode == http.StatusOK || statusCode == http.StatusCreated || statusCode == http.StatusAccepted || statusCode == http.StatusNoContent {
		return nil
	}

	return HandleError(httpResponseBody)
}

func HandleError(httpResponseBody []byte) error {
	var apiError numspot.Error

	err := json.Unmarshal(httpResponseBody, &apiError)
	if err != nil {
		return errors.New("API Error : " + string(httpResponseBody))
	}

	errorString := apiError.Title
	if apiError.Detail != nil && *apiError.Detail != "" {
		errorString = errorString + ": " + *apiError.Detail
	}

	return errors.New(errorString)
}
