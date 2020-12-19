package service

import (
	"bytes"
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"net/url"
)

// DelegateAuthService provide verification of token by checking with controller
type DelegateAuthService struct {
	controllerOrigin string
}

// NewDelegateAuthService create new auth service connecting to specified controlller path
func NewDelegateAuthService(controllerOrigin string) *DelegateAuthService {
	return &DelegateAuthService{
		controllerOrigin: controllerOrigin,
	}
}

// Verify check with controller and return userId of token
// UserID is return only err != nil
func (auth *DelegateAuthService) Verify(token string) (string, error) {
	data, err := json.Marshal(map[string]interface{}{
		"token": token,
	})
	if err != nil {
		return "", err
	}
	endpoint := url.URL{
		Scheme: "http",
		Host:   auth.controllerOrigin,
		Path:   "/api/v1/user/verify",
	}
	resp, err := http.Post(endpoint.String(), "application/json", bytes.NewReader(data))
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	resBody, err := ioutil.ReadAll(resp.Body)

	if err != nil {
		return "", err
	}

	if resp.StatusCode == http.StatusOK {
		var resOK struct {
			UserID string `json:"userId"`
		}
		err = json.Unmarshal(resBody, &resOK)
		if err != nil {
			return "", err
		}

		return resOK.UserID, nil
	} else {
		var resError struct {
			Status string `json:"status"`
		}
		err = json.Unmarshal(resBody, &resError)
		if err != nil {
			return "", errors.New(string(resBody))
		}

		return "", errors.New(resError.Status)
	}
}
