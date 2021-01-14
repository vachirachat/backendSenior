package encryption

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"errors"
	"fmt"
	"io"

	"github.com/mergermarket/go-pkcs7"
)

// source: https://gist.github.com/brettscott/2ac58ab7cb1c66e2b4a32d6c1c3908a7

// AESEncrypt encrypt message with key
func AESEncrypt(plainText []byte, key []byte) ([]byte, error) {
	if key == nil {
		return nil, errors.New("key is nil")
	}

	plainText, err := pkcs7.Pad(plainText, aes.BlockSize)
	if err != nil {
		return nil, err
	}

	if len(plainText)%aes.BlockSize != 0 {
		return nil, fmt.Errorf("padding error: wrong size")
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	cipherText := make([]byte, len(plainText)+aes.BlockSize)
	iv := cipherText[:aes.BlockSize]
	_, err = io.ReadFull(rand.Reader, iv)
	if err != nil {
		return nil, fmt.Errorf("error init iv: %s", err.Error())
	}

	mode := cipher.NewCBCEncrypter(block, iv)
	mode.CryptBlocks(cipherText[aes.BlockSize:], plainText)

	return cipherText, nil
}

// AESDecrypt decrypt message with key
func AESDecrypt(cipherText []byte, key []byte) ([]byte, error) {
	if key == nil {
		return nil, errors.New("key is nil")
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	if len(cipherText) < aes.BlockSize {
		return nil, errors.New("cipher text too short")
	}

	iv := cipherText[:aes.BlockSize]
	data := cipherText[aes.BlockSize:]
	if len(data)%aes.BlockSize != 0 {
		return nil, errors.New("wrong cipher text size")
	}

	mode := cipher.NewCBCDecrypter(block, iv)
	if err != nil {
		return nil, err
	}

	decrypted := make([]byte, len(data))
	mode.CryptBlocks(decrypted, data)

	if !isValid(decrypted) {
		return nil, errors.New("corrupted message, wrong key?")
	}
	decrypted, _ = pkcs7.Unpad(decrypted, aes.BlockSize)

	return decrypted, nil
}

// isValid is used for checking whether decrypted is valid before unpad
func isValid(padded []byte) bool {
	// this is taken for pkcs7 source code
	bufLen := len(padded) - int(padded[len(padded)-1])
	if bufLen < 0 {
		return false
	}
	return true
}
