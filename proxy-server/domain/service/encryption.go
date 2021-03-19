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

	"github.com/mergermarket/go-pkcs7"

	"proxySenior/domain/interface/repository"
	"proxySenior/domain/plugin"
)

// EncryptionService is service for encrpyting and decrypting message
type EncryptionService struct {
	keystore repository.Keystore
	plugin   *plugin.OnMessagePortPlugin
}

// NewEncryptionService create instance of encryption service
func NewEncryptionService(keystore repository.Keystore, onMessagePortPlugin *plugin.OnMessagePortPlugin) *EncryptionService {
	return &EncryptionService{
		keystore: keystore,
		plugin:   onMessagePortPlugin,
	}
}

// source: https://gist.github.com/brettscott/2ac58ab7cb1c66e2b4a32d6c1c3908a7

func (enc *EncryptionService) EncryptController(message model.Message) (model.Message, error) {

	if enc.plugin.IsEnabledEncryption() {
		log.Println("Test IsEnabledEncryption Select", "True")
		return enc.plugin.CustomEncryptionPlugin(message)
	} else {
		log.Println("Test Select", "False")
		return enc.EncryptBase(message)
	}
}

func (enc *EncryptionService) DecryptController(message model.Message) (model.Message, error) {
	if enc.plugin.IsEnabledEncryption() {
		log.Println("Test DecryptController Select", "True")
		return enc.plugin.CustomDecryptionPlugin(message)
	} else {
		log.Println("Test Select", "False")
		return enc.DecryptBase(message)
	}
}

// Encrypt takes a message, then return message with data encrypted
// Task: Plugin-Encryption :: Use As base Encryption
func (enc *EncryptionService) EncryptBase(message model.Message) (model.Message, error) {
	keyRec, err := enc.keystore.GetKeyForMessage(message.RoomID.Hex(), message.TimeStamp)
	if err != nil {
		return message, fmt.Errorf("getting key: %s", err.Error())
	}

	key := keyRec.Key
	plainText := []byte(message.Data)
	plainText, err = pkcs7.Pad(plainText, aes.BlockSize)
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

// Decrypt takes a message, then return message with data decrypted with appropiate key
func (enc *EncryptionService) DecryptBase(message model.Message) (model.Message, error) {
	// fmt.Printf("[Decode] original message text is [%s]\n", message.Data)
	// decoder := base64.NewDecoder(base64.StdEncoding, bytes.NewReader([]byte(message.Data)))
	// decoded, err := ioutil.ReadAll(decoder)
	// if err != nil {
	// 	message.Data = "Error Decoding: " + err.Error()
	// } else {
	// 	message.Data = string(decoded)
	// }
	keyRec, err := enc.keystore.GetKeyForMessage(message.RoomID.Hex(), message.TimeStamp)
	if err != nil {
		return message, fmt.Errorf("getting key: %s", err.Error())
	}

	key := keyRec.Key

	// b64 := base64.NewDecoder(base64.StdEncoding, bytes.NewReader([]byte(message.Data)))
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
