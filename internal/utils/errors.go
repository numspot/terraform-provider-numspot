package utils

import (
	"encoding/json"
	"fmt"
	"net/http"
	"reflect"
	"regexp"
	"runtime"
	"strings"

	"github.com/getkin/kin-openapi/openapi3filter"
	"github.com/hashicorp/terraform-plugin-framework/diag"
)

type NSError struct {
	Type    string `json:"Type"`
	Details string `json:"Details"`
	Code    string `json:"Code"`
}

type ApiError struct {
	Errors          []NSError `json:"Errors"`
	ResponseContext struct {
		RequestID string `json:"RequestId"`
	} `json:"ResponseContext"`
}

func (a ApiError) Type() string {
	return a.Errors[0].Type
}

func (a ApiError) Error() string {
	if len(a.Errors) > 0 {
		return a.Errors[0].Details
	}
	return ""
}

func HandleError(httpResponseBody []byte) error {
	apiError := ApiError{}
	err := json.Unmarshal(httpResponseBody, &apiError)
	if err != nil {
		apiError.Errors = append(apiError.Errors, NSError{Details: string(httpResponseBody)})
	}

	return apiError
}

func getCallerFunctionName() string {
	// Profondeur de 2 pour obtenir la fonction qui a appelé la fonction actuelle
	pc, file, line, ok := runtime.Caller(2)

	fmt.Println(file, line, ok)

	// Obtient le nom de la fonction en utilisant la réflexion
	fn := runtime.FuncForPC(pc)
	if fn == nil {
		return "Inconnue"
	}
	return fn.Name()
}

func ExecuteRequest[A openapi3filter.StatusCoder](fun func() (*A, error), expectedStatusCode int, diagnostics *diag.Diagnostics) *A {
	res, err := fun()
	if err != nil {
		diagnostics.AddError("Failed", err.Error())
		return nil
	}

	// This is employed for code reduction; however, using reflection is discouraged
	// to remove when go integrate proposal to handle multiple struct access ...
	statusCode, body := reflectHttpResponse(*res)
	if statusCode == nil {
		diagnostics.AddError("HTTP Response error", "Failed to reflect http response")
		return nil
	}

	if expectedStatusCode != *statusCode {
		stack := getCallerFunctionName()
		split := strings.Split(stack, ".")
		structName := split[len(split)-2]
		operationName := strings.ToLower(split[len(split)-1])

		r := regexp.MustCompile(`\w`)
		matches := r.FindAllString(structName, -1)
		structName = strings.Join(matches, "")
		structName = strings.ReplaceAll(structName, "Resource", "")

		apiError := HandleError(body)
		diagnostics.AddError(fmt.Sprintf("Failed to %s %s", operationName, structName), apiError.Error())
		return nil
	}

	return res
}

func reflectHttpResponse(res openapi3filter.StatusCoder) (statusCode *int, bodyBytes []byte) {
	value := reflect.ValueOf(res)
	if value.Kind() == reflect.Pointer {
		value = value.Elem()
	}

	for i := 0; i < value.NumField(); i++ {
		typeField := value.Type().Field(i)

		if typeField.Name == "HTTPResponse" {
			valueField := value.Field(i)
			httpResponse, ok := valueField.Interface().(*http.Response)
			if ok {
				statusCode = &httpResponse.StatusCode
			}
		} else if typeField.Name == "Body" {
			valueField := value.Field(i)
			bytes, ok := valueField.Interface().([]byte)
			if ok {
				bodyBytes = bytes
			}
		}
	}

	return
}
