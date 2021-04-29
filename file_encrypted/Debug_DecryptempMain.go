package main

// "github.com/alexmullins/zip"

// var key = "0123456789abcdef"

// func main() {

// 	// List of Files to Zip
// 	// files := []string{"text_1.txt", "text_2.txt"}
// 	filename := "to_zip_text_1.txt"
// 	output := "done.zip"
// 	dat, err := ioutil.ReadFile(filename)
// 	if err != nil {
// 		panic(err)
// 	}
// 	log.Println("File Data EncryptBase>>", string(dat), "\n")
// 	fileData, err := DecryptBase(string(dat), key)
// 	if err != nil {
// 		panic(err)
// 	}

// 	log.Println("File Data DecryptBase>>", fileData, "\n")
// 	fmt.Println("Zipped File:", output)

// }

// func EncryptBase(stringFile string, key string) (string, error) {
// 	plainText := []byte(stringFile)
// 	plainText, err := pkcs7.Pad(plainText, aes.BlockSize)
// 	if err != nil {
// 		return "", fmt.Errorf("padding plaintext: %s", err.Error())
// 	}

// 	if len(plainText)%aes.BlockSize != 0 {
// 		return "", fmt.Errorf("padding error: wrong size")
// 	}

// 	block, err := aes.NewCipher([]byte(key))
// 	if err != nil {
// 		return "", err
// 	}

// 	cipherText := make([]byte, len(plainText)+aes.BlockSize)
// 	iv := cipherText[:aes.BlockSize]
// 	_, err = io.ReadFull(rand.Reader, iv)
// 	if err != nil {
// 		return "", fmt.Errorf("error init iv: %s", err.Error())
// 	}

// 	mode := cipher.NewCBCEncrypter(block, iv)
// 	mode.CryptBlocks(cipherText[aes.BlockSize:], plainText)

// 	// encode base 64 before send
// 	var result bytes.Buffer
// 	b64 := base64.NewEncoder(base64.StdEncoding, &result)
// 	b64.Write(cipherText)
// 	b64.Close()

// 	return result.String(), nil
// }

// // Decrypt takes a message, then return message with data decrypted with appropiate key
// func DecryptBase(stringFile string, key string) (string, error) {
// 	cipherText, err := base64.StdEncoding.DecodeString(stringFile)
// 	if err != nil {
// 		return "", fmt.Errorf("decode b64: %s", err.Error())
// 	}

// 	block, err := aes.NewCipher([]byte(key))
// 	if err != nil {
// 		return "", err
// 	}

// 	if len(cipherText) < aes.BlockSize {
// 		return "", errors.New("cipher text too short")
// 	}

// 	iv := cipherText[:aes.BlockSize]
// 	data := cipherText[aes.BlockSize:]
// 	if len(data)%aes.BlockSize != 0 {
// 		return "", errors.New("wrong cipher text size")
// 	}

// 	mode := cipher.NewCBCDecrypter(block, iv)
// 	if err != nil {
// 		return "", err
// 	}

// 	decrypted := make([]byte, len(data))

// 	mode.CryptBlocks(decrypted, data)

// 	decrypted, _ = pkcs7.Unpad(decrypted, aes.BlockSize)

// 	return string(decrypted), nil
// }
