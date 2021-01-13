package delegate

import (
	"backendSenior/domain/model/chatsocket/key_exchange"
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"proxySenior/domain/interface/repository"
)

type KeyAPI struct {
	origin string // host:port of controller
}

// KeyAPI implement KeyStore (read part only btw)
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
	if res.StatusCode >= 400 {
		return key_exchange.KeyExchangeResponse{}, fmt.Errorf("server return with non OK status: %d", res.StatusCode)
	}

	body, err = ioutil.ReadAll(res.Body)
	defer res.Body.Close()

	var keys key_exchange.KeyExchangeResponse
	err = json.Unmarshal(body, &keys)
	if err != nil {
		return key_exchange.KeyExchangeResponse{}, fmt.Errorf("error decoding response: %v", err)
	}
	fmt.Printf("[get key] response :%s\n", body)
	return keys, nil
}
