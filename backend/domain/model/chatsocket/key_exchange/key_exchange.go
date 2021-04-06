package key_exchange

import (
	model_proxy "proxySenior/domain/model"
)

// KeyExchangeRequest is request for key
type KeyExchangeRequest struct {
	// [optional] public key for other side to encrypt
	// if not specified will use cached key
	PublicKey []byte `json:"publicKey,omitempty"`
	// ProxyID that is target of request, used by controller to route request
	ProxyID string `json:"proxyId,omitempty"`
	// RoomID roomID to request for key
	RoomID string `json:"roomId,omitempty"`
}

// KeyExchangeResponse is response containing key in exchange proess
type KeyExchangeResponse struct {
	// [optional] public key for other side to encrypt
	// if not specified will use cached key
	PublicKey []byte `json:"publicKey,omitempty"`
	// ProxyID requester of this response, used by controller to route request
	ProxyID string `json:"proxyId,omitempty"`
	// RoomID roomID for this response
	RoomID string
	// Keys keys of the room, data of the key shall be encrypted
	Keys []model_proxy.KeyRecord
	// ErrorMessage explaining error
	ErrorMessage string
}

// ToReadable return readable representation
func (r *KeyExchangeResponse) ToReadable() map[string]interface{} {

	return map[string]interface{}{
		"publicKey": string(r.PublicKey), // it's simply byte, not base64 encoded
		"proxyId":   r.ProxyID,
		"roomId":    r.RoomID,
		"keys":      nil,
	}
}
