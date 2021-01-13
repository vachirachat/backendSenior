package service

import (
	"backendSenior/domain/model/chatsocket/key_exchange"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha512"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"errors"
	"fmt"
	"log"
	"os"
	"proxySenior/domain/interface/repository"
	model_proxy "proxySenior/domain/model"
	"time"

	"github.com/globalsign/mgo/bson"
)

// KeyService is service for managing key
type KeyService struct {
	local    repository.Keystore // local
	remote   repository.RemoteKeyStore
	proxy    repository.ProxyMasterAPI
	clientID string // clientID of this proxy
}

// NewKeyService create new key service
func NewKeyService(local repository.Keystore, remote repository.RemoteKeyStore, proxy repository.ProxyMasterAPI, clientID string) *KeyService {
	return &KeyService{
		local:    local,
		remote:   remote,
		proxy:    proxy,
		clientID: clientID,
	}
}

// ---- local

// GetKeyLocal is used for getting keys for room locally if possible
func (s *KeyService) GetKeyLocal(roomID string) ([]model_proxy.KeyRecord, error) {
	ok, err := s.IsLocal(roomID)
	if err != nil {
		return nil, fmt.Errorf("error checking locality of key: %v", err)
	}

	if ok {
		keys, err := s.local.Find(model_proxy.KeyRecordUpdate{
			RoomID: bson.ObjectIdHex(roomID),
		})
		if err == nil && keys == nil {
			keys = []model_proxy.KeyRecord{}
		}
		return keys, err
	}

	return nil, errors.New("can't get local key for remote room")
}

// NewKeyForRoom generate new key for room, invalidating old one
func (s *KeyService) NewKeyForRoom(roomID string) error {
	if ok, err := s.IsLocal(roomID); err != nil {
		return fmt.Errorf("error checking locality of key: %v", err)
	} else if !ok {
		return errors.New("can't generate key for remote proxy")
	}

	key, err := randomBytes(32)
	if err != nil {
		return err
	}

	err = s.local.AddNewKey(roomID, key)

	return err
}

type isLocalEntry struct {
	data    bool
	expires time.Time
}

var isLocalCache = make(map[string]isLocalEntry)

// IsLocal return whether key from `roomID` can be fetched locally (by key store)
func (s *KeyService) IsLocal(roomID string) (bool, error) {
	if cache, ok := isLocalCache[roomID]; ok {
		if cache.expires.After(time.Now()) {
			return cache.data, nil
		}
	}

	proxy, err := s.proxy.GetRoomMasterProxy(roomID)
	if err != nil {
		return false, err
	}
	isLocal := proxy.ProxyID.Hex() == s.clientID
	isLocalCache[roomID] = isLocalEntry{
		data:    isLocal,
		expires: time.Now().Add(1 * time.Minute),
	}
	return isLocal, nil
}

type keyRemoteEntry struct {
	data    []model_proxy.KeyRecord
	expires time.Time
}

var keyRemoteCache = make(map[string]keyRemoteEntry)

// GetKeyRemote is used to get key from remote
func (s *KeyService) GetKeyRemote(roomID string) ([]model_proxy.KeyRecord, error) {
	if cache, ok := keyRemoteCache[roomID]; ok {
		if cache.expires.After(time.Now()) {
			return cache.data, nil
		}
	}

	proxy, err := s.proxy.GetRoomMasterProxy(roomID)
	if err != nil {
		return nil, fmt.Errorf("can't determine master proxy: %v", err)
	}

	// TODO: currently only used for checking
	// whether we had communicate w/ this proxy before
	// if not we want to send PK
	// their PL isn't used anyway (for now)
	var reqBody *key_exchange.KeyExchangeRequest
	_, ok := cache[proxy.ProxyID.Hex()]

	// TODO refactor whole file
	myID := os.Getenv("CLIENT_ID")

	if !ok {
		fmt.Println("request with key")
		reqBody = &key_exchange.KeyExchangeRequest{
			PublicKey: s.PublicKeyToBytes(s.GetMyPublicKey()),
			ProxyID:   myID,
			RoomID:    roomID,
		}
	} else {
		fmt.Println("request without key")
		reqBody = &key_exchange.KeyExchangeRequest{
			PublicKey: nil,
			ProxyID:   myID,
			RoomID:    roomID,
		}
	}

	resp, err := s.remote.GetByRoom(roomID, *reqBody)
	if err != nil {
		return nil, err
	}

	// we had request w/ public key, so extract the PK from response
	if !ok {
		s.SetProxyPublicKey(proxy.ProxyID.Hex(), resp.PublicKey)
	}

	// decrypt the kesy
	for i := range resp.Keys {
		dec := s.DecryptWithPrivateKey(resp.Keys[i].Key, s.GetMyPrivateKey())
		resp.Keys[i].Key = dec
	}

	_respJSON, _ := json.Marshal(resp) // so we can see byte message easier
	fmt.Printf("[get-key-remote] roomId: %s\ndecrypted keys: %s\n", roomID, _respJSON)

	keyRemoteCache[roomID] = keyRemoteEntry{
		data:    resp.Keys,
		expires: time.Now().Add(10 * time.Second),
	}

	return resp.Keys, nil
}

// TODO: use other way to generate
// generate key, size should be 32
func randomBytes(size int) ([]byte, error) {
	key := make([]byte, size)
	n, err := rand.Read(key)
	if err != nil || n != size {
		return nil, err
	}
	return key, err
}

// public and private keys

var public *rsa.PublicKey
var private *rsa.PrivateKey
var cache = make(map[string]*rsa.PublicKey)

// InitKeyPair create key pair
func (s *KeyService) InitKeyPair() {
	private, public = s.generateKeyPair(2048)
}

// GetMyPrivateKey
func (s *KeyService) GetMyPrivateKey() *rsa.PrivateKey {
	return private
}

func (s *KeyService) GetMyPublicKey() *rsa.PublicKey {
	return public
}

func (s *KeyService) GetProxyPublicKey(proxyID string) (*rsa.PublicKey, bool) {
	key, ok := cache[proxyID]
	return key, ok
}

// SetProxyPublicKey key should be byte sent from
func (s *KeyService) SetProxyPublicKey(proxyID string, key []byte) {
	pk := s.BytesToPublicKey(key)
	if pk == nil {
		log.Println("set pk error")
	}
	cache[proxyID] = pk
}

// PK utils

// EncryptWithPublicKey encrypts data with public key
func (s *KeyService) EncryptWithPublicKey(msg []byte, pub *rsa.PublicKey) []byte {
	hash := sha512.New()
	ciphertext, err := rsa.EncryptOAEP(hash, rand.Reader, pub, msg, nil)
	if err != nil {
		log.Println(err)
	}
	return ciphertext
}

// DecryptWithPrivateKey decrypts data with private key
func (s *KeyService) DecryptWithPrivateKey(ciphertext []byte, priv *rsa.PrivateKey) []byte {
	hash := sha512.New()
	plaintext, err := rsa.DecryptOAEP(hash, rand.Reader, priv, ciphertext, nil)
	if err != nil {
		log.Println(err)
	}
	return plaintext
}

// generateKeyPair generates a new key pair
func (s *KeyService) generateKeyPair(bits int) (*rsa.PrivateKey, *rsa.PublicKey) {
	privkey, err := rsa.GenerateKey(rand.Reader, bits)
	if err != nil {
		log.Println(err)
	}
	return privkey, &privkey.PublicKey
}

// PrivateKeyToBytes private key to bytes
func (s *KeyService) PrivateKeyToBytes(priv *rsa.PrivateKey) []byte {
	privBytes := pem.EncodeToMemory(
		&pem.Block{
			Type:  "RSA PRIVATE KEY",
			Bytes: x509.MarshalPKCS1PrivateKey(priv),
		},
	)

	return privBytes
}

// PublicKeyToBytes public key to bytes
func (s *KeyService) PublicKeyToBytes(pub *rsa.PublicKey) []byte {
	pubASN1, err := x509.MarshalPKIXPublicKey(pub)
	if err != nil {
		log.Println(err)
	}

	pubBytes := pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PUBLIC KEY",
		Bytes: pubASN1,
	})

	return pubBytes
}

// BytesToPrivateKey bytes to private key
func (s *KeyService) BytesToPrivateKey(priv []byte) *rsa.PrivateKey {
	block, _ := pem.Decode(priv)
	enc := x509.IsEncryptedPEMBlock(block)
	b := block.Bytes
	var err error
	if enc {
		log.Println("is encrypted pem block")
		b, err = x509.DecryptPEMBlock(block, nil)
		if err != nil {
			log.Println(err)
		}
	}
	key, err := x509.ParsePKCS1PrivateKey(b)
	if err != nil {
		log.Println(err)
	}
	return key
}

// BytesToPublicKey bytes to public key
func (s *KeyService) BytesToPublicKey(pub []byte) *rsa.PublicKey {
	block, _ := pem.Decode(pub)
	enc := x509.IsEncryptedPEMBlock(block)
	b := block.Bytes
	var err error
	if enc {
		log.Println("is encrypted pem block")
		b, err = x509.DecryptPEMBlock(block, nil)
		if err != nil {
			log.Println(err)
		}
	}
	ifc, err := x509.ParsePKIXPublicKey(b)
	if err != nil {
		log.Println(err)
	}
	key, ok := ifc.(*rsa.PublicKey)
	if !ok {
		log.Println("not ok")
	}
	return key
}
