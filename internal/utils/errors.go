package utils

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"terraform-provider-numspot/internal/sdk/api"
)

// errAPIError is the static sentinel for API error responses.
var errAPIError = errors.New("API error")

// ParseHTTPError handles parse HTTP error.
func ParseHTTPError(httpResponseBody []byte, statusCode int) error {
	if statusCode == http.StatusOK || statusCode == http.StatusCreated || statusCode == http.StatusAccepted || statusCode == http.StatusNoContent {
		return nil
	}

	return HandleError(httpResponseBody)
}

// HandleError handles handle error.
func HandleError(httpResponseBody []byte) error {
	var apiError api.Error

	err := json.Unmarshal(httpResponseBody, &apiError)
	if err != nil && len(httpResponseBody) != 0 {
		return fmt.Errorf("%w: %s", errAPIError, string(httpResponseBody))
	}

	errorString := apiError.Title
	if apiError.Detail != nil && *apiError.Detail != "" {
		errorString = errorString + ": " + *apiError.Detail
	}

	return fmt.Errorf("%w: %s", errAPIError, errorString)
}
