package service

import (
	"backendSenior/domain/model"
	"bytes"
	"encoding/base64"
	"fmt"
	"io/ioutil"
)

type EncryptionService struct {
	// TODO add something like keystore
}

// Encrypt takes a message, then return message with data encrypted
func (enc *EncryptionService) Encrypt(message model.Message) model.Message {
	// TODO: encryption logic
	var buff bytes.Buffer
	encoder := base64.NewEncoder(base64.StdEncoding, &buff)
	encoder.Write([]byte(message.Data))
	encoder.Close()
	message.Data = buff.String()
	return message
}

// Decrypt takes a message, then return message with data decrypted with appropiate key
func (enc *EncryptionService) Decrypt(message model.Message) model.Message {
	// TODO: encryption logic
	fmt.Printf("[Decode] original message text is [%s]\n", message.Data)
	decoder := base64.NewDecoder(base64.StdEncoding, bytes.NewReader([]byte(message.Data)))
	decoded, err := ioutil.ReadAll(decoder)
	if err != nil {
		message.Data = "Error Decoding: " + err.Error()
	} else {
		message.Data = string(decoded)
	}
	return message
}
