package conns

import (
	"encoding/json"
	"fmt"
	"net/http"
)

func HandleError(expectedStatusCode int, response *http.Response) *NumSpotError {
	if expectedStatusCode != response.StatusCode {
		var numspotError NumSpotError
		defer response.Body.Close()
		err := json.NewDecoder(response.Body).Decode(&numspotError)
		fmt.Println(err)

		return &numspotError
	}

	return nil
}
