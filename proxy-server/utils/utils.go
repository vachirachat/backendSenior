package utils

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os/exec"
	"proxySenior/config"

	docker "github.com/fsouza/go-dockerclient"
	"github.com/mergermarket/go-pkcs7"
)

func DecrytedFile(fileName string) error {
	dat, err := ioutil.ReadFile(config.PATH_ORIGIN_ZIP + fileName)
	if err != nil {
		return err
	}
	// Function Decrypted
	fileData, err := DecryptBase(string(dat), config.PROXY_KEY)
	if err != nil {
		return err
	}

	err = ioutil.WriteFile(config.PATH_ORIGIN_ZIP+"en_"+fileName, []byte(fileData), 0644)
	if err != nil {
		return err
	}
	return nil
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
	// decrypted := make([]byte, aes.BlockSize)

	mode.CryptBlocks(decrypted, data)

	decrypted, _ = pkcs7.Unpad(decrypted, aes.BlockSize)

	return string(decrypted), nil
}

// Decrypt takes a message, then return message with data decrypted with appropiate key
func DecryptBaseCode(stringCode string, keys string) (string, error) {
	key := []byte(keys)
	cipherText, _ := hex.DecodeString(stringCode)

	block, err := aes.NewCipher(key)
	if err != nil {
		panic(err)
	}

	if len(cipherText) < aes.BlockSize {
		panic("cipherText too short")
	}
	iv := cipherText[:aes.BlockSize]
	cipherText = cipherText[aes.BlockSize:]
	if len(cipherText)%aes.BlockSize != 0 {
		panic("cipherText is not a multiple of the block size")
	}

	mode := cipher.NewCBCDecrypter(block, iv)
	mode.CryptBlocks(cipherText, cipherText)

	cipherText, _ = pkcs7.Unpad(cipherText, aes.BlockSize)
	return fmt.Sprintf("%s", cipherText), nil
}

func UnzipFile(fileName string) error {
	cmdUnzip := exec.Command("unzip", config.PATH_ORIGIN+fileName, "-d", config.PATH_ORIGIN_ZIP)
	_, err := cmdUnzip.Output()

	if err != nil {
		log.Println("Cannot UnzipFile")
		return err
	}
	return nil
}

// helper docker
func RunDockerExec(containerID string, workingDir string, cmdInput []string) (string, error) {
	// dockerAPIversion, _ := docker.NewAPIVersion()
	// dockerClient, err := docker.NewClientFromEnv()
	apiVersionString := "1.41"
	dockerClient, err := docker.NewVersionedClientFromEnv(apiVersionString)
	if err != nil {
		log.Panicln(">> ", err)
		return "", err
	}
	dockerClient.SkipServerVersionCheck = apiVersionString == ""

	if err != nil {
		log.Panicln("Error: %s", err)
		return "", err
	}

	dockerExOpt := docker.CreateExecOptions{
		AttachStderr: true,
		AttachStdin:  true,
		AttachStdout: true,
		Tty:          false,
		WorkingDir:   workingDir,
		Cmd:          cmdInput,
		Container:    containerID,
	}
	dockerExec, err := dockerClient.CreateExec(dockerExOpt)
	if err != nil {
		return "", err
	}
	// Start Execfile
	var stdout, stderr bytes.Buffer
	execID := dockerExec.ID
	opts := docker.StartExecOptions{
		OutputStream: &stdout,
		ErrorStream:  &stderr,
		RawTerminal:  true,
	}

	err = dockerClient.StartExec(execID, opts)
	if err != nil {
		return "", err
	}
	// log.Println("stdout: %+s", stdout.String())
	return stdout.String(), err

}

// helper pipeline call
func Execute(output_buffer *bytes.Buffer, stack ...*exec.Cmd) (err error) {
	var error_buffer bytes.Buffer
	pipe_stack := make([]*io.PipeWriter, len(stack)-1)
	i := 0
	for ; i < len(stack)-1; i++ {
		stdin_pipe, stdout_pipe := io.Pipe()
		stack[i].Stdout = stdout_pipe
		stack[i].Stderr = &error_buffer
		stack[i+1].Stdin = stdin_pipe
		pipe_stack[i] = stdout_pipe
	}
	stack[i].Stdout = output_buffer
	stack[i].Stderr = &error_buffer

	if err := call(stack, pipe_stack); err != nil {
		log.Println(string(error_buffer.Bytes()), err)
	}
	return err
}

// helper execute pipeline call
func call(stack []*exec.Cmd, pipes []*io.PipeWriter) (err error) {
	if stack[0].Process == nil {
		if err = stack[0].Start(); err != nil {
			return err
		}
	}
	if len(stack) > 1 {
		if err = stack[1].Start(); err != nil {
			return err
		}
		defer func() {
			if err == nil {
				pipes[0].Close()
				err = call(stack[1:], pipes[1:])
			}
		}()
	}
	return stack[0].Wait()
}
