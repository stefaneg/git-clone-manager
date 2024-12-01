package gitlab

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	. "tools/internal/log"
)

func gitlabGet[T any](token string, url string) (T, error) {
	var emptyResult T
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return emptyResult, err
	}
	req.Header.Set("PRIVATE-TOKEN", token)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return emptyResult, err
	}
	defer func(body io.ReadCloser) {
		err := body.Close()
		if err != nil {
			Log.Errorf("Failed to close response body: %v", err)
		}
	}(resp.Body)

	if resp.StatusCode != http.StatusOK {
		return emptyResult, fmt.Errorf("GitLab API request on %s failed with status: %s", url, resp.Status)
	}

	var decodedResult T
	if err := json.NewDecoder(resp.Body).Decode(&decodedResult); err != nil {
		return emptyResult, err
	}

	return decodedResult, nil
}
