package delegate

import (
	"backendSenior/domain/model/chatsocket/key_exchange"
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"proxySenior/config"
	"proxySenior/domain/interface/repository"
)

// KeyAPI is for getting key remotely
type KeyAPI struct {
	origin string // host:port of controller
}

// KeyAPI implement RemoteKeyStore
var _ repository.RemoteKeyStore = (*KeyAPI)(nil)

// NewKeyAPI is api for contacting controller
func NewKeyAPI(controller string) *KeyAPI {
	return &KeyAPI{
		origin: controller,
	}
}

// GetByRoom ask key for `roomID` from controller
func (a *KeyAPI) GetByRoom(roomID string, details key_exchange.KeyExchangeRequest) (key_exchange.KeyExchangeResponse, error) {
	fmt.Println("[remote key store] get key for room", roomID)
	u := url.URL{
		Scheme: "http",
		Host:   a.origin,
		Path:   "/api/v1/key/room-key/" + roomID,
	}

	body, err := json.Marshal(details)
	bodyReader := bytes.NewReader(body)

	res, err := http.Post(u.String(), "appliation/json", bodyReader)
	if err != nil {
		return key_exchange.KeyExchangeResponse{}, fmt.Errorf("error making request: %v", err)
	}

	// event with non-OK status, we still want to return the response
	isOK := true
	if res.StatusCode >= 400 {
		isOK = false
	}

	body, err = ioutil.ReadAll(res.Body)
	defer res.Body.Close()

	var dataResp key_exchange.KeyExchangeResponse
	err = json.Unmarshal(body, &dataResp)
	if err != nil {
		return key_exchange.KeyExchangeResponse{}, fmt.Errorf("error decoding response: %v", err)
	}
	// fmt.Printf("[get key] response OK?=%v :%s\n", isOK, body)

	// force error message
	if !isOK {
		err = fmt.Errorf("server return with non OK status: %d\nbody:%s", res.StatusCode, body)
	}
	return dataResp, err
}

func (a *KeyAPI) CatchUp(roomID string) error {
	fmt.Println("[remote key store] report catchup", roomID)
	u := url.URL{
		Scheme: "http",
		Host:   a.origin,
		Path:   "/api/v1/key/catch-up/" + roomID + "/" + config.ClientID,
	}

	res, err := http.Post(u.String(), "appliation/json", nil)
	if err != nil {
		return fmt.Errorf("error making request: %v", err)
	}

	// event with non-OK status, we still want to return the response
	isOK := true
	if res.StatusCode >= 400 {
		isOK = false
	}

	body, err := ioutil.ReadAll(res.Body)
	defer res.Body.Close()

	// force error message
	if !isOK {
		err = fmt.Errorf("server return with non OK status: %d\nbody:%s", res.StatusCode, body)
	}
	return nil
}
