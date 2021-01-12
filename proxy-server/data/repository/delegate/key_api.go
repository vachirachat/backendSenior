package delegate

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"proxySenior/domain/interface/repository"
	model_proxy "proxySenior/domain/model"
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
func (a *KeyAPI) GetByRoom(roomID string) ([]model_proxy.KeyRecord, error) {
	fmt.Println("[remote key store] get key for room", roomID)
	u := url.URL{
		Scheme: "http",
		Host:   a.origin,
		Path:   "/api/v1/key/room-key/" + roomID,
	}
	defer print("error if not see result")

	res, err := http.Get(u.String())
	if err != nil {
		return nil, fmt.Errorf("error making request: %v", err)
	}
	if res.StatusCode >= 400 {
		return nil, fmt.Errorf("server return with non OK status: %d", res.StatusCode)
	}

	body, err := ioutil.ReadAll(res.Body)
	fmt.Printf("body: %s\n", body)
	defer res.Body.Close()

	var keys []model_proxy.KeyRecord
	err = json.Unmarshal(body, &keys)
	if err != nil {
		return nil, fmt.Errorf("error decoding response: %v", err)
	}
	fmt.Println("[remote key store] result ", keys)
	return keys, nil
}
