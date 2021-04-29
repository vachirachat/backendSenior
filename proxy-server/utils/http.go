package utils

import (
	"encoding/base64"
	"fmt"
	"proxySenior/config"
)

type BadStatusError struct {
	Status int
	Body   []byte
}

// Error implements error interface
func (e BadStatusError) Error() string {
	return fmt.Sprint("server returned with non-OK status:", e.Status)
}

//
//// HTTPGet perform HTTP ger then unmarshal response to result
//// It consider response an error if status >= 400
//func HTTPGet(url string, result interface{}) error {
//	res, err := http.Get(url)
//	if err != nil {
//		return fmt.Errorf("request: %w", err)
//	}
//
//	body, err := ioutil.ReadAll(res.Body)
//	defer res.Body.Close()
//
//	if res.StatusCode >= 400 {
//		return BadStatusError{
//			Status: res.StatusCode,
//			Body:   body,
//		}
//	}
//
//	err = json.Unmarshal(body, &result)
//	if err != nil {
//		return fmt.Errorf("unmarshal: %w", err)
//	}
//	return nil
//}
//
//// HTTPPost perform HTTP post then unmarshal response to result
//// It consider response an error if status >= 400
//func HTTPPost(url string, contentType string, data interface{}, result interface{}) error {
//	dataJSON, err := json.Marshal(data)
//	if err != nil {
//		return fmt.Errorf("marshal error %w", err)
//	}
//
//	res, err := http.Post(url, contentType, bytes.NewReader(dataJSON))
//	if err != nil {
//		return fmt.Errorf("request: %w", err)
//	}
//
//	body, err := ioutil.ReadAll(res.Body)
//	defer res.Body.Close()
//
//	if res.StatusCode >= 400 {
//		return BadStatusError{
//			Status: res.StatusCode,
//			Body:   body,
//		}
//	}
//
//	err = json.Unmarshal(body, &result)
//	if err != nil {
//		return fmt.Errorf("unmarshal: %w", err)
//	}
//	return nil
//}

// AuthHeader return header for basic aauth
func AuthHeader() string {
	return "Basic " + base64.StdEncoding.EncodeToString([]byte(config.ClientID+":"+config.ClientSecret))
}
