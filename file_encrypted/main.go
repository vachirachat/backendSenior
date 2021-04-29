package main

import (
	"archive/zip"
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"

	"github.com/mergermarket/go-pkcs7"
	// "github.com/alexmullins/zip"
)

var key = "0123456789abcdef"

func main() {

	// List of Files to Zip
	// files := []string{"text_1.txt", "text_2.txt"}
	files := []string{"docker_exec"}
	output := "done.zip"

	if err := ZipFiles(output, files); err != nil {
		panic(err)
	}
	fmt.Println("Zipped File:", output)

}

// ZipFiles compresses one or many files into a single zip archive file.
// Param 1: filename is the output zip file's name.
// Param 2: files is a list of files to add to the zip.
func ZipFiles(filename string, files []string) error {

	newZipFile, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer newZipFile.Close()

	zipWriter := zip.NewWriter(newZipFile)

	// Add files to zip
	for _, file := range files {
		if err = AddFileToZip(zipWriter, file); err != nil {
			return err
		}
	}
	defer zipWriter.Close()
	return nil
}

func AddFileToZip(zipWriter *zip.Writer, filename string) error {

	dat, err := ioutil.ReadFile(filename)
	if err != nil {
		return err
	}
	// log.Println("File Data>>", string(dat), "\n")
	fileData, err := EncryptBase(string(dat), key)
	if err != nil {
		return err
	}
	err = ioutil.WriteFile("en_"+filename, []byte(fileData), 0644)
	if err != nil {
		return err
	}

	fileToZip, err := os.Open("en_" + filename)
	if err != nil {
		return err
	}
	defer fileToZip.Close()

	// Get the file information
	info, err := fileToZip.Stat()
	if err != nil {
		return err
	}

	header, err := zip.FileInfoHeader(info)
	if err != nil {
		return err
	}

	// Using FileInfoHeader() above only uses the basename of the file. If we want
	// to preserve the folder structure we can overwrite this with the full path.
	header.Name = filename
	header.Comment = "Secret"
	// log.Println("header FileInfo", header.FileInfo())
	// log.Println("\n header ", header)
	// Change to deflate to gain better compression
	// see http://golang.org/pkg/archive/zip/#pkg-constants
	header.Method = zip.Deflate

	writer, err := zipWriter.CreateHeader(header)
	if err != nil {
		return err
	}

	// Test loop
	// v := reflect.ValueOf(*header)
	// typeOfS := v.Type()

	// for i := 0; i < v.NumField(); i++ {
	// 	fmt.Printf("Field: %s\tValue: %v\n", typeOfS.Field(i).Name, v.Field(i).Interface())
	// }

	_, err = io.Copy(writer, fileToZip)
	return err
}

func EncryptBase(stringFile string, key string) (string, error) {
	plainText := []byte(stringFile)
	plainText, err := pkcs7.Pad(plainText, aes.BlockSize)
	if err != nil {
		return "", fmt.Errorf("padding plaintext: %s", err.Error())
	}

	if len(plainText)%aes.BlockSize != 0 {
		return "", fmt.Errorf("padding error: wrong size")
	}

	block, err := aes.NewCipher([]byte(key))
	if err != nil {
		return "", err
	}

	cipherText := make([]byte, len(plainText)+aes.BlockSize)
	iv := cipherText[:aes.BlockSize]
	_, err = io.ReadFull(rand.Reader, iv)
	if err != nil {
		return "", fmt.Errorf("error init iv: %s", err.Error())
	}

	mode := cipher.NewCBCEncrypter(block, iv)
	mode.CryptBlocks(cipherText[aes.BlockSize:], plainText)

	// encode base 64 before send
	var result bytes.Buffer
	b64 := base64.NewEncoder(base64.StdEncoding, &result)
	b64.Write(cipherText)
	b64.Close()

	return result.String(), nil
}

// Decrypt takes a message, then return message with data decrypted with appropiate key
func DecryptBase(stringFile string, key string) (string, error) {
	cipherText, err := base64.StdEncoding.DecodeString(stringFile)
	if err != nil {
		return "", fmt.Errorf("decode b64: %s", err.Error())
	}

	block, err := aes.NewCipher([]byte(key))
	if err != nil {
		return "", err
	}

	if len(cipherText) < aes.BlockSize {
		return "", errors.New("cipher text too short")
	}

	iv := cipherText[:aes.BlockSize]
	data := cipherText[aes.BlockSize:]
	if len(data)%aes.BlockSize != 0 {
		return "", errors.New("wrong cipher text size")
	}

	mode := cipher.NewCBCDecrypter(block, iv)
	if err != nil {
		return "", err
	}

	decrypted := make([]byte, len(data))

	mode.CryptBlocks(decrypted, data)

	decrypted, _ = pkcs7.Unpad(decrypted, aes.BlockSize)

	return string(decrypted), nil
}
