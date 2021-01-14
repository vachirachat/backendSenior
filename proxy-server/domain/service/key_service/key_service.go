package key_service

import (
	"backendSenior/domain/model/chatsocket/key_exchange"
	"crypto/rand"
	"crypto/rsa"
	"encoding/json"
	"fmt"
	"proxySenior/domain/encryption"
	"proxySenior/domain/interface/repository"
	model_proxy "proxySenior/domain/model"
	"proxySenior/utils"
	"time"
)

// KeyService is service for managing key
// - manage symmetric key generation and query locally
// - manage getting symmetric key from
type KeyService struct {
	local        repository.Keystore // local
	remote       repository.RemoteKeyStore
	proxy        repository.ProxyMasterAPI
	clientID     string // clientID of this proxy
	keyCache     map[string]keyCacheEntry
	pubKeyCache  map[string]*rsa.PublicKey
	isLocalCache map[string]isLocalEntry
	public       *rsa.PublicKey
	privateKey   *rsa.PrivateKey
}

// keyCacheEntry is used for caching key with expire time
type keyCacheEntry struct {
	data    []model_proxy.KeyRecord // keys
	expires time.Time               // cache expires
}

// isLocalEntry is used for caching answer for `IsLocal`
type isLocalEntry struct {
	data    bool      // whether room is local
	expires time.Time // cache expires
}

// New create new key service
func New(local repository.Keystore, remote repository.RemoteKeyStore, proxy repository.ProxyMasterAPI, clientID string) *KeyService {
	return &KeyService{
		local:    local,
		remote:   remote,
		proxy:    proxy,
		clientID: clientID,
		// cache
		keyCache:     make(map[string]keyCacheEntry),
		pubKeyCache:  make(map[string]*rsa.PublicKey),
		isLocalCache: make(map[string]isLocalEntry),
		// keypair
		public:     nil,
		privateKey: nil,
	}
}

// GetKeyRemote get room-key from room remotely.
// it determine proxy from room automatically.
// sendKey determine whether we will additionally exchange public key
func (s *KeyService) GetKeyRemote(roomID string) ([]model_proxy.KeyRecord, error) {
	// memoization
	if cache, ok := s.keyCache[roomID]; ok {
		if cache.expires.After(time.Now()) {
			return cache.data, nil
		}
	}

	proxy, err := s.proxy.GetRoomMasterProxy(roomID)
	if err != nil {
		return nil, fmt.Errorf("can't determine master proxy: %v", err)
	}

	// get past proxy key for the proxy
	_, ok := s.pubKeyCache[proxy.ProxyID.Hex()]

	// key not exists (ok) then send key
	resp, err := s.getRoomKey(roomID, !ok)

	if err != nil {
		// if otherside have lost the key, the send key gain
		if resp.ErrorMessage == "ERR_NO_KEY" {
			resp, err = s.getRoomKey(roomID, true)
			if err != nil {
				return nil, err
			}
		} else {
			return nil, err
		}
	}

	// now we get key
	err = decryptRespKeys(&resp, s.privateKey)
	// if decrypt error, maybe pub key changed, we try again
	if err != nil {
		resp, err = s.getRoomKey(roomID, true)
		if err != nil {
			return nil, err
		}
		err = decryptRespKeys(&resp, s.privateKey)
		if err != nil {
			return nil, err
		}
	}

	_respJSON, _ := json.Marshal(resp) // so we can see byte message easier
	fmt.Printf("[get-key-remote] roomId: %s\ndecrypted keys: %s\n", roomID, _respJSON)

	s.keyCache[roomID] = keyCacheEntry{
		data:    resp.Keys,
		expires: time.Now().Add(10 * time.Second),
	}

	return resp.Keys, nil
}

// decryptRespKeys modifies resp to decrypt the key, it
func decryptRespKeys(resp *key_exchange.KeyExchangeResponse, privKey *rsa.PrivateKey) error {
	for i := range resp.Keys {
		decryptedKey, err := encryption.DecryptWithPrivateKey(resp.Keys[i].Key, privKey)
		if err != nil {
			return err
		}
		resp.Keys[i].Key = decryptedKey
	}
	return nil
}

// InitKeyPair create key pair
// It should be called before using any function
func (s *KeyService) InitKeyPair() {
	s.ensureSelfKeys()
}

// GetProxyPublicKey return cached public key for proxy
func (s *KeyService) GetProxyPublicKey(proxyID string) (*rsa.PublicKey, bool) {
	key, ok := s.pubKeyCache[proxyID]
	return key, ok
}

// SetProxyPublicKey set public key to cache
func (s *KeyService) SetProxyPublicKey(proxyID string, key *rsa.PublicKey) {
	s.pubKeyCache[proxyID] = key
}

// PK utils

// generateKeyPair generates a new key pair
func generateKeyPair(bits int) (*rsa.PrivateKey, *rsa.PublicKey) {
	privkey, err := rsa.GenerateKey(rand.Reader, bits)
	if err != nil {
		panic(fmt.Sprintf("can't generate key: %v", err))
	}
	return privkey, &privkey.PublicKey
}

func (s *KeyService) ensureSelfKeys() {
	// already generated
	if s.public != nil {
		return
	}
	priv, pub := generateKeyPair(2048)
	s.privateKey = priv
	s.public = pub
}

// getRoomKey is helper for API of getting room key with or without key
// the key responed is NOT decrypted
func (s *KeyService) getRoomKey(roomID string, sendKey bool) (key_exchange.KeyExchangeResponse, error) {
	var reqBody *key_exchange.KeyExchangeRequest
	pubkey := s.public
	if pubkey == nil {
		return key_exchange.KeyExchangeResponse{}, fmt.Errorf("self public key is nil")
	}

	if sendKey {
		fmt.Println("[keyservice] get room key with public key")
		reqBody = &key_exchange.KeyExchangeRequest{
			PublicKey: encryption.PublicKeyToBytes(s.public),
			ProxyID:   utils.ClientID,
			RoomID:    roomID,
		}
	} else {
		fmt.Println("[keyservice] get room key WITHOUT public key")
		reqBody = &key_exchange.KeyExchangeRequest{
			PublicKey: nil,
			ProxyID:   utils.ClientID,
			RoomID:    roomID,
		}
	}

	resp, err := s.remote.GetByRoom(roomID, *reqBody)
	return resp, err
}

// generate key, size should be 32
func randomBytes(size int) ([]byte, error) {
	key := make([]byte, size)
	n, err := rand.Read(key)
	if err != nil || n != size {
		return nil, err
	}
	return key, err
}
