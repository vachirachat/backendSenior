package utils

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
)

type BadStatusError struct {
	Status int
	Body   []byte
}

// Error implements error interface
func (e BadStatusError) Error() string {
	return fmt.Sprint("server returned with non-OK status:", e.Status)
}

// HTTPGet perform HTTP ger then unmarshal response to result
// It consider response an error if status >= 400
func HTTPGet(u url.URL, result interface{}) error {
	res, err := http.Get(u.String())
	if err != nil {
		return fmt.Errorf("request: %w", err)
	}

	body, err := ioutil.ReadAll(res.Body)
	defer res.Body.Close()

	if res.StatusCode >= 400 {
		return BadStatusError{
			Status: res.StatusCode,
			Body:   body,
		}
	}

	err = json.Unmarshal(body, &result)
	if err != nil {
		return fmt.Errorf("unmarshal: %w", err)
	}
	return nil
}
