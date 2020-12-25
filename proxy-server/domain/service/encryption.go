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
	log.Println("\n Encrypted log.Println(keyRec.KeyRecodes --> \n")
	keyRec, err := enc.keystore.GetKeyForMessage(message.RoomID.Hex(), message.TimeStamp)

	if err != nil {
		keyRec, err = enc.keystore.AddNewKey(message.RoomID.Hex(), keyRec.KeyRecodes)
		if err != nil {
			return message, fmt.Errorf("Fail to Add Key: %s", err.Error())
		}
	} else {
		keyRec, err = enc.keystore.UpdateNewKey(message.RoomID.Hex(), keyRec.KeyRecodes, message.TimeStamp)
		if err != nil {
			return message, fmt.Errorf("Fail to Update Key: %s", err.Error())
		}
	}

	//For Test Propose
	for i := range keyRec.KeyRecodes {
		log.Print(keyRec.KeyRecodes[i].Key)
	}

	key := keyRec.KeyRecodes[len(keyRec.KeyRecodes)-1].Key

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
func (enc *EncryptionService) Decrypt(message model.Message) (model.Message, error) {
	log.Println("\n Decrypted log.Println(keyRec.KeyRecodes --> \n")
	// fmt.Printf("[Decode] original message text is [%s]\n", message.Data)
	// decoder := base64.NewDecoder(base64.StdEncoding, bytes.NewReader([]byte(message.Data)))
	// decoded, err := ioutil.ReadAll(decoder)
	// if err != nil {
	// 	message.Data = "Error Decoding: " + err.Error()
	// } else {
	// 	message.Data = string(decoded)
	// }
	keyRec, err := enc.keystore.GetKeyForMessage(message.RoomID.Hex(), message.TimeStamp)
	//----- Refactor to Function Check key-date -----//

	if err != nil {
		return message, fmt.Errorf("getting key: %s", err.Error())
	}
	//----- TODO :: Refactor to Function -----//
	//----- Refactor to Function ReMapTime -----//
	loc, err := time.LoadLocation(utils.BACKKOKTIMEZONE)
	key := keyRec.KeyRecodes[0].Key
	for i := range keyRec.KeyRecodes {
		keyMap := keyRec.KeyRecodes[i]
		keyMap.ValidTo = keyMap.ValidTo.In(loc)
		keyMap.ValidFrom = keyMap.ValidFrom.In(loc)
		if keyMap.ValidTo.After(message.TimeStamp.In(loc)) && keyMap.ValidFrom.Before(message.TimeStamp.In(loc)) {
			key = keyMap.Key
			log.Print("Match key >>> ", key)
			continue
		}
	}
	//----- Refactor to Function ReMapTime -----//

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
