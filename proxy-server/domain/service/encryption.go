package service

import (
	"backendSenior/domain/model"
	"fmt"
	"proxySenior/domain/encryption"
	model_proxy "proxySenior/domain/model"
	"proxySenior/domain/service/key_service"
	"time"

	"proxySenior/domain/plugin"
)

// EncryptionService is service for encrpyting and decrypting message
type EncryptionService struct {
	key    *key_service.KeyService
	plugin *plugin.OnMessagePortPlugin
}

// NewEncryptionService create instance of encryption service
func NewEncryptionService(key *key_service.KeyService, onMessagePortPlugin *plugin.OnMessagePortPlugin) *EncryptionService {
	return &EncryptionService{
		key:    key,
		plugin: onMessagePortPlugin,
	}
}

// source: https://gist.github.com/brettscott/2ac58ab7cb1c66e2b4a32d6c1c3908a7
func (enc *EncryptionService) EncryptController(msg *model.Message) error {

	if enc.plugin.IsEnabledEncryption() {

		decrypted, err := enc.plugin.CustomEncryptionPlugin(*msg)
		if err != nil {
			return fmt.Errorf("plugin encrypt: %w", err)
		}
		msg.Data = decrypted.Data
		return nil

	} else {
		//log.Println("Test Select", "False")
		return enc.encryptBase(msg)
	}
}

func (enc *EncryptionService) DecryptController(msg *model.Message) error {
	if enc.plugin.IsEnabledEncryption() {
		//log.Println("Test DecryptController Select", "True")
		decrypted, err := enc.plugin.CustomDecryptionPlugin(*msg)
		if err != nil {
			return fmt.Errorf("plugin decrypt: %w", err)
		}
		msg.Data = decrypted.Data
		return nil

	} else {
		//log.Println("Test Select", "False")
		return enc.decryptBase(msg)
	}
}

// encryptBase encrypt message using standard logic
func (enc *EncryptionService) encryptBase(msg *model.Message) error {
	// get key
	keys, err := enc.getKeyFromRoom(msg.RoomID.Hex())
	if err != nil {
		return fmt.Errorf("getting key for room: %w", err)
	}
	now := time.Now()
	key := keyFor(keys, now)

	// encrypt and encode afterwards
	encrypted, err := encryption.AESEncrypt([]byte(msg.Data), key)
	if err != nil {
		return fmt.Errorf("encrypting message: %w", err)
	}
	encrypted = encryption.B64Encode(encrypted)

	msg.Data = string(encrypted)
	return nil
}

// decryptBase decrypt message using standard logic
func (enc *EncryptionService) decryptBase(msg *model.Message) error {
	// get key
	keys, err := enc.getKeyFromRoom(msg.RoomID.Hex())
	if err != nil {
		return fmt.Errorf("getting key for room: %w", err)
	}
	key := keyFor(keys, msg.TimeStamp)

	// encrypt and encode afterwards
	cipherText, err := encryption.B64Decode([]byte(msg.Data))
	if err != nil {
		return fmt.Errorf("b64decode error: %w", err)
	}
	decrypted, err := encryption.AESDecrypt(cipherText, key)
	if err != nil {
		return fmt.Errorf("encrypting message: %w", err)
	}
	msg.Data = string(decrypted)
	return nil
}

// getKeyFromRoom helper function for retreiving key for
func (enc *EncryptionService) getKeyFromRoom(roomID string) ([]model_proxy.KeyRecord, error) {
	local, err := enc.key.IsLocal(roomID)
	if err != nil {
		return nil, fmt.Errorf("error deftermining locality ok key %v", err)
	}

	var keys []model_proxy.KeyRecord
	if local {
		//fmt.Println("[message] use LOCAL key for", roomID)
		keys, err = enc.key.GetKeyLocal(roomID)
		if err != nil {
			return nil, fmt.Errorf("error getting key locally %v", err)
		}
	} else {
		//fmt.Println("[message] use REMOTE key for room", roomID)
		keys, err = enc.key.GetKeyRemote(roomID)
		if err != nil {
			return nil, fmt.Errorf("error getting key remotely %v", err)
		}
	}

	return keys, nil
}

// keyFor is helper function for finding key in array by time
func keyFor(keys []model_proxy.KeyRecord, timestamp time.Time) []byte {
	var key []byte
	found := false
	for _, k := range keys {
		if k.ValidFrom.Before(timestamp) && (k.ValidTo.IsZero() || k.ValidTo.After(timestamp)) {
			key = k.Key
			found = true
			break
		}
	}
	if !found {
		return nil
	}
	return key
}
