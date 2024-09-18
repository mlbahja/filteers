package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
)

// fetchAndDecode performs an HTTP GET request to the specified URL and decodes the JSON response into the provided target variable.
func fetchAndDecode(url string, target interface{}) error {
	// Perform the HTTP GET request to the specified URL
	response, err := http.Get(url)
	if err != nil {
		// Return an error if the request fails
		return fmt.Errorf("error fetching data: %w", err)
	}
	// Ensure that the response body is closed once the function exits
	defer response.Body.Close()

	// Decode the JSON response into the target variable
	err = json.NewDecoder(response.Body).Decode(target)
	if err != nil {
		// Return an error if decoding fails
		return fmt.Errorf("error decoding response: %w", err)
	}

	// Return nil if there were no errors
	return nil
}
