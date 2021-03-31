package key_service

import (
	"backendSenior/domain/model/chatsocket/key_exchange"
	"crypto/rand"
	"crypto/rsa"
	"fmt"
	"github.com/cornelk/hashmap"
	"proxySenior/domain/encryption"
	"proxySenior/domain/interface/repository"
	model_proxy "proxySenior/domain/model"
	"proxySenior/utils"
	"strings"
)

// KeyService is service for managing key
// - manage symmetric key generation and query locally
// - manage getting symmetric key from
type KeyService struct {
	local           repository.Keystore // local
	remote          repository.RemoteKeyStore
	proxy           repository.ProxyMasterAPI
	clientID        string // clientID of this proxy
	keyCache        *hashmap.HashMap
	pubKeyCache     *hashmap.HashMap
	roomMasterCache *hashmap.HashMap
	public          *rsa.PublicKey
	privateKey      *rsa.PrivateKey
}

// New create new key service
func New(local repository.Keystore, remote repository.RemoteKeyStore, proxy repository.ProxyMasterAPI, clientID string) *KeyService {
	return &KeyService{
		local:    local,
		remote:   remote,
		proxy:    proxy,
		clientID: clientID,
		// cache
		keyCache:        hashmap.New(32), // cache room key make(map[string][]model_proxy.KeyRecord)
		pubKeyCache:     hashmap.New(32), // make(map[string]*rsa.PublicKey)          // cache proxy public key
		roomMasterCache: hashmap.New(32), // make(map[string]model.Proxy),             // ca
		// keypair
		public:     nil,
		privateKey: nil,
	}
}

const (
	badStatusErrorMessage = "server return with non OK status"
)

// GetKeyRemote get room-key from room remotely.
// it determine proxy from room automatically.
// sendKey determine whether we will additionally exchange public key
func (s *KeyService) GetKeyRemote(roomID string) ([]model_proxy.KeyRecord, error) {
	// memoization
	if keys, ok := s.keyCache.Get(roomID); ok {
		fmt.Println("[REMOTE] cached key")
		return keys.([]model_proxy.KeyRecord), nil
	}

	proxy, err := s.proxy.GetRoomMasterProxy(roomID)
	if err != nil {
		return nil, fmt.Errorf("can't determine master proxy: %v", err)
	}

	// get past proxy key for the proxy
	_, ok := s.pubKeyCache.Get(proxy.ProxyID.Hex())

	// key not exists (ok) then send key

	resp, err := s.getRoomKey(roomID, !ok)
	// fmt.Println("response is", resp)
	if err != nil {
		// if otherside have lost the key, the send key gain
		if strings.HasPrefix(err.Error(), badStatusErrorMessage) && resp.ErrorMessage == "ERR_NO_KEY" {
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

	// if success we cache the key to DB and report to controller
	err = s.local.ReplaceKey(roomID, resp.Keys)
	if err != nil {
		fmt.Println("update key error:", err)
	}
	s.remote.CatchUp(roomID)

	s.keyCache.Set(roomID, resp.Keys)

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
	key, ok := s.pubKeyCache.Get(proxyID)
	return key.(*rsa.PublicKey), ok
}

// SetProxyPublicKey set public key to cache
func (s *KeyService) SetProxyPublicKey(proxyID string, key *rsa.PublicKey) {
	s.pubKeyCache.Set(proxyID, key)
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
