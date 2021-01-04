package service

import (
	"backendSenior/domain/model"
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"log"
	model_proxy "proxySenior/domain/model"
	"time"

	"github.com/mergermarket/go-pkcs7"

	"proxySenior/domain/interface/repository"
	"proxySenior/utils"
)

// EncryptionService is service for encrpyting and decrypting message
type EncryptionService struct {
	keystore repository.Keystore
}

// NewEncryptionService create instance of encryption service
func NewEncryptionService(keystore repository.Keystore) *EncryptionService {
	return &EncryptionService{
		keystore: keystore,
	}
}

// source: https://gist.github.com/brettscott/2ac58ab7cb1c66e2b4a32d6c1c3908a7

// Encrypt takes a message, then return message with data encrypted
func (enc *EncryptionService) Encrypt(message model.Message) (model.Message, error) {
	keyRec, err := enc.keystore.GetKeyForMessage(message.RoomID.Hex(), message.TimeStamp)

	if err != nil {
		keyRec, err = enc.keystore.AddNewKey(message.RoomID.Hex(), keyRec.KeyRecodes)
		if err != nil {
			return message, fmt.Errorf("Fail to Add Key: %s", err.Error())
		}
	}

	loc, err := time.LoadLocation(utils.BACKKOKTIMEZONE)
	keyTime, err := validateTime(keyRec.KeyRecodes[len(keyRec.KeyRecodes)-1])
	if err != nil {
		return message, fmt.Errorf("Fail to Add Change TimeZone: %s", err.Error())
	}

	if keyTime.ValidTo.Before(time.Now().In(loc)) {
		keyRec, err = enc.keystore.UpdateNewKey(message.RoomID.Hex(), keyRec.KeyRecodes, message.TimeStamp)
		if err != nil {
			return message, fmt.Errorf("Fail to Update Key: %s", err.Error())
		}
	}
	return encryptedMessage(message, keyRec.KeyRecodes[len(keyRec.KeyRecodes)-1].Key)
}

// Decrypt takes a message, then return message with data decrypted with appropiate key
func (enc *EncryptionService) Decrypt(message model.Message) (model.Message, error) {
	keyRec, err := enc.keystore.GetKeyForMessage(message.RoomID.Hex(), message.TimeStamp)
	if err != nil {
		return message, fmt.Errorf("getting key: %s", err.Error())
	}
	key, err := decryptedKeyValidate(keyRec, message)
	if err != nil {
		return message, fmt.Errorf("Fail to Decrypt: %s", err.Error())
	}
	return decryptedMessage(message, key)
}

func validateTime(keyRec model_proxy.KeyRecord) (model_proxy.KeyRecord, error) {
	loc, err := time.LoadLocation(utils.BACKKOKTIMEZONE)
	keyRec.ValidTo = keyRec.ValidTo.In(loc)
	keyRec.ValidFrom = keyRec.ValidFrom.In(loc)
	return keyRec, err
}

// validate key by time
func decryptedKeyValidate(keyRec model_proxy.RoomKeys, message model.Message) ([]byte, error) {
	loc, err := time.LoadLocation(utils.BACKKOKTIMEZONE)
	key := keyRec.KeyRecodes[0].Key
	for i := range keyRec.KeyRecodes {
		keyMap, err := validateTime(keyRec.KeyRecodes[i])
		if err != nil {
			return make([]byte, 0), fmt.Errorf("Fail to Add Change TimeZone: %s", err.Error())
		}
		if keyMap.ValidTo.After(message.TimeStamp.In(loc)) && keyMap.ValidFrom.Before(message.TimeStamp.In(loc)) {
			key = keyMap.Key
			log.Print("Match key >>> ", key)
			return key, nil
		}
	}
	return make([]byte, 0), fmt.Errorf("Key not Found : %s", err.Error())
}

func decryptedMessage(message model.Message, key []byte) (model.Message, error) {
	cipherText, err := base64.StdEncoding.DecodeString(message.Data)
	if err != nil {
		return message, fmt.Errorf("decode b64: %s", err.Error())
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return message, err
	}

	if len(cipherText) < aes.BlockSize {
		return message, errors.New("cipher text too short")
	}

	iv := cipherText[:aes.BlockSize]
	data := cipherText[aes.BlockSize:]
	if len(data)%aes.BlockSize != 0 {
		return message, errors.New("wrong cipher text size")
	}

	mode := cipher.NewCBCDecrypter(block, iv)
	if err != nil {
		return message, err
	}

	decrypted := make([]byte, len(data))

	mode.CryptBlocks(decrypted, data)

	decrypted, _ = pkcs7.Unpad(decrypted, aes.BlockSize)
	message.Data = string(decrypted)
	return message, nil
}

func encryptedMessage(message model.Message, key []byte) (model.Message, error) {
	plainText := []byte(message.Data)
	plainText, err := pkcs7.Pad(plainText, aes.BlockSize)
	if err != nil {
		return message, fmt.Errorf("padding plaintext: %s", err.Error())
	}

	if len(plainText)%aes.BlockSize != 0 {
		return message, fmt.Errorf("padding error: wrong size")
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return message, err
	}

	cipherText := make([]byte, len(plainText)+aes.BlockSize)
	iv := cipherText[:aes.BlockSize]
	_, err = io.ReadFull(rand.Reader, iv)
	if err != nil {
		return message, fmt.Errorf("error init iv: %s", err.Error())
	}

	mode := cipher.NewCBCEncrypter(block, iv)
	mode.CryptBlocks(cipherText[aes.BlockSize:], plainText)

	// encode base 64 before send
	var result bytes.Buffer
	b64 := base64.NewEncoder(base64.StdEncoding, &result)
	b64.Write(cipherText)
	b64.Close()

	message.Data = result.String()
	return message, nil
}
