package service

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"os"
	"os/exec"
	model_proxy "proxySenior/domain/model"
	"proxySenior/domain/plugin"
	"proxySenior/utils"
	"strings"
)

// MessageService i sservice for getting message from controller and decrypt it for user
type ConfigService struct {
	// configRepo  repository.ConfigRepository
	onMessagePortPlugin *plugin.OnMessagePortPlugin
	proxyConfig         *model_proxy.ProxyConfig
	encryption          *EncryptionService
}

// NewMessageService create new instance of message service
// func NewConfigService(configRepo repository.ConfigRepository, enc *EncryptionService, proxyConfig model_proxy.ProxyConfig) *ConfigService {
func NewConfigService(enc *EncryptionService, proxyConfig *model_proxy.ProxyConfig, onMessagePortPlugin *plugin.OnMessagePortPlugin) *ConfigService {
	configService := &ConfigService{
		// configRepo:  configRepo,
		onMessagePortPlugin: onMessagePortPlugin,
		proxyConfig:         proxyConfig,
		encryption:          enc,
	}
	// Create DockerImage when-start proxy
	if proxyConfig.EnablePlugin {
		configService.startDockerImage()
	}
	return configService

}

func (confService *ConfigService) ConfigGetPluginStatus() bool {
	return confService.proxyConfig.EnablePlugin
}

func (confService *ConfigService) ConfigSetStartProxy() {
	confService.proxyConfig.EnablePlugin = true
	return
}

func (confService *ConfigService) ConfigSetStopProxy() {
	confService.proxyConfig.EnablePlugin = false
	return
}

func (confService *ConfigService) ConfigFileProxy(file io.Reader, fileHandler *multipart.FileHeader) error {
	fmt.Printf("Uploaded File: %+v\n", fileHandler.Filename)
	fmt.Printf("File Size: %+v\n", fileHandler.Size)
	fmt.Printf("MIME Header: %+v\n", fileHandler.Header)

	// StartDockerImage
	confService.startDockerImage()
	// Create file
	dst, err := os.Create(utils.PATH_ORIGIN + fileHandler.Filename)
	defer dst.Close()
	if err != nil {
		return err
	}

	// Copy the uploaded file to the created file on the filesystem
	if _, err := io.Copy(dst, file); err != nil {
		return err
	}

	err = utils.UnzipFile(fileHandler.Filename)
	if err != nil {
		return err
	}

	err = utils.DecrytedFile(utils.DOCKEREXEC_FILE_NAME)
	if err != nil {
		return err
	}

	// cmdChmod := exec.Command("chmod", "+x", utils.PATH_ORIGIN_ZIP+"to_zip_"+fileHandler.Filename)
	cmdChmod := exec.Command("chmod", "+x", utils.PATH_ORIGIN_ZIP+"to_zip_"+utils.DOCKEREXEC_FILE_NAME)
	log.Println("chmod", "+x", utils.PATH_ORIGIN_ZIP+"to_zip_"+utils.DOCKEREXEC_FILE_NAME)
	_, err = cmdChmod.Output()
	if err != nil {
		log.Println("Cannot change file-type")
		log.Fatal(err)
		return err
	}
	log.Println("docker", "cp", utils.PATH_ORIGIN_ZIP+"to_zip_"+utils.DOCKEREXEC_FILE_NAME, confService.proxyConfig.DockerID+":"+utils.DOCKER_PATH_ORIGIN+"/exec-module")
	// cmdDockerCopy := exec.Command("docker", "cp", utils.PATH_ORIGIN_ZIP+"to_zip_"+fileHandler.Filename, *confService.proxyConfig.DockerID+":"+utils.DOCKER_PATH_ORIGIN+"/exec-module")
	cmdDockerCopy := exec.Command("docker", "cp", utils.PATH_ORIGIN_ZIP+"to_zip_"+utils.DOCKEREXEC_FILE_NAME, confService.proxyConfig.DockerID+":"+utils.DOCKER_PATH_ORIGIN+"/exec-module")

	_, err = cmdDockerCopy.Output()
	if err != nil {
		log.Println("Cannot cmdDockerCopy")
		log.Fatal(err)
		return err
	}

	// var b bytes.Buffer
	// if err = utils.Execute(&b,
	// 	exec.Command("rm", "-Rf", utils.PATH_ORIGIN+"*"),
	// 	exec.Command("yes"),
	// ); err != nil {
	// 	return err
	// }

	// if err = utils.Execute(&b,
	// 	exec.Command("rm", "-Rf", utils.PATH_ORIGIN_ZIP+"*"),
	// 	exec.Command("yes"),
	// ); err != nil {
	// 	return err
	// }

	cmdRemove := exec.Command("rm", utils.PATH_ORIGIN+fileHandler.Filename)
	_, err = cmdRemove.Output()
	if err != nil {
		log.Println("Cannot remove PATH_ORIGIN file-type")
		log.Fatal(err)
	}

	cmdRemove = exec.Command("rm", utils.PATH_ORIGIN_ZIP+utils.DOCKEREXEC_FILE_NAME)
	_, err = cmdRemove.Output()
	if err != nil {
		log.Println("Cannot remove PATH_ORIGIN_ZIP file-type")
		log.Fatal(err)
	}

	cmdRemove = exec.Command("rm", utils.PATH_ORIGIN_ZIP+"to_zip_"+utils.DOCKEREXEC_FILE_NAME)
	_, err = cmdRemove.Output()
	if err != nil {
		log.Println("Cannot remove PATH_ORIGIN_ZIP file-type")
		log.Fatal(err)
	}

	return nil
}

func (confService *ConfigService) ConfigRunCodeProxy(storage model_proxy.JSONCODE) error {
	if storage.Lang == "go" {
<<<<<<< HEAD
		out, err := utils.RunDockerExec(confService.proxyConfig.DockerID, "/app/go_server/"+storage.Lang+"-module/", []string{"go", "run", storage.Filename + "." + storage.Lang})
		log.Println("out file go Run >>>>", out)
		return err
	} else if storage.Lang == "py" {
		out, err := utils.RunDockerExec(confService.proxyConfig.DockerID, "/app/go_server/"+storage.Lang+"-module/", []string{"python3", storage.Filename + "." + storage.Lang})
=======
		out, err := utills.RunDockerExec(confService.proxyConfig.DockerID, "/app/go_server/"+storage.Lang+"-module/", []string{"go", "run", storage.Filename + "." + storage.Lang})
		log.Println("out file go Run >>>>", out)
		return err
	} else if storage.Lang == "py" {
		out, err := utills.RunDockerExec(confService.proxyConfig.DockerID, "/app/go_server/"+storage.Lang+"-module/", []string{"python3", storage.Filename + "." + storage.Lang})
>>>>>>> feat/proxy/code-api
		log.Println("out file py Run >>>>", out)
		return err
		// _ = runDockerExec(containerID, []string{"python3", "run", storage.Lang + "-module/file." + storage.Lang, "&"})
	} else if storage.Lang == "js" {
<<<<<<< HEAD
		out, err := utils.RunDockerExec(confService.proxyConfig.DockerID, "/app/go_server/"+storage.Lang+"-module/", []string{"node", storage.Filename + "." + storage.Lang})
=======
		out, err := utills.RunDockerExec(confService.proxyConfig.DockerID, "/app/go_server/"+storage.Lang+"-module/", []string{"node", storage.Filename + "." + storage.Lang})
>>>>>>> feat/proxy/code-api
		log.Println("out file js Run >>>>", out)
		return err

	}
	return nil
}

var key = os.Getenv("PLUGIN_Encryption_Key")

func (confService *ConfigService) ConfigCodeProxy(storage model_proxy.JSONCODE) error {
	// StartDockerImage
	confService.startDockerImage()
	// Create file

	log.Println("Lang", storage.Lang)
	log.Println("Code", storage.Code)
<<<<<<< HEAD
	storage.Code, _ = utils.DecryptBaseCode(storage.Code, key)
	// File Upload
	f, err := os.Create(utils.PATH_ORIGIN_PROXY + "/" + storage.Lang + "/file." + storage.Lang)
=======
	storage.Code, _ = utills.DecryptBaseCode(storage.Code, key)
	// File Upload
	f, err := os.Create(utills.PATH_ORIGIN_PROXY + "/" + storage.Lang + "/file." + storage.Lang)
>>>>>>> feat/proxy/code-api
	if err != nil {
		log.Println("ConfigCodeProxy Create file", err)
		return err

	}
	defer f.Close()
	w := bufio.NewWriter(f)
	n4, err := w.WriteString(storage.Code)
	if err != nil {
		log.Println("ConfigCodeProxy Copy data to file", err)
		return err

	}
	fmt.Printf("wrote %d bytes\n", n4)
	w.Flush()

	// Upload to Docker
<<<<<<< HEAD
	log.Println("docker", "cp", utils.PATH_ORIGIN_PROXY+"/"+storage.Lang+"/file."+storage.Lang, confService.proxyConfig.DockerID+":"+utils.DOCKER_PATH_ORIGIN+"/go-module")
	if storage.Lang == "go" {
		// Copy to docker
		cmdCovert := exec.Command("docker", "cp", utils.PATH_ORIGIN_PROXY+"/"+storage.Lang+"/file."+storage.Lang, confService.proxyConfig.DockerID+":"+utils.DOCKER_PATH_ORIGIN+"/go-module")
=======
	log.Println("docker", "cp", utills.PATH_ORIGIN_PROXY+"/"+storage.Lang+"/file."+storage.Lang, confService.proxyConfig.DockerID+":"+utills.DOCKER_PATH_ORIGIN+"/go-module")
	if storage.Lang == "go" {
		// Copy to docker
		cmdCovert := exec.Command("docker", "cp", utills.PATH_ORIGIN_PROXY+"/"+storage.Lang+"/file."+storage.Lang, confService.proxyConfig.DockerID+":"+utills.DOCKER_PATH_ORIGIN+"/go-module")
>>>>>>> feat/proxy/code-api
		_, err = cmdCovert.Output()
		if err != nil {
			log.Println("Cannot cmdCovert go")
			return err
		}
	} else if storage.Lang == "py" {
<<<<<<< HEAD
		cmdCovert := exec.Command("docker", "cp", utils.PATH_ORIGIN_PROXY+"/"+storage.Lang+"/file."+storage.Lang, confService.proxyConfig.DockerID+":"+utils.DOCKER_PATH_ORIGIN+"/py-module")
=======
		cmdCovert := exec.Command("docker", "cp", utills.PATH_ORIGIN_PROXY+"/"+storage.Lang+"/file."+storage.Lang, confService.proxyConfig.DockerID+":"+utills.DOCKER_PATH_ORIGIN+"/py-module")
>>>>>>> feat/proxy/code-api
		_, err = cmdCovert.Output()
		if err != nil {
			log.Println("Cannot cmdCovert py")
			return err
		}
	} else if storage.Lang == "js" {
<<<<<<< HEAD
		cmdCovert := exec.Command("docker", "cp", utils.PATH_ORIGIN_PROXY+"/"+storage.Lang+"/file."+storage.Lang, confService.proxyConfig.DockerID+":"+utils.DOCKER_PATH_ORIGIN+"/js-module")
=======
		cmdCovert := exec.Command("docker", "cp", utills.PATH_ORIGIN_PROXY+"/"+storage.Lang+"/file."+storage.Lang, confService.proxyConfig.DockerID+":"+utills.DOCKER_PATH_ORIGIN+"/js-module")
>>>>>>> feat/proxy/code-api
		_, err = cmdCovert.Output()
		if err != nil {
			log.Println("Cannot cmdCovert js")
			return err
		}
	}
	// Run code when upload sucess
	err = confService.ConfigRunCodeProxy(model_proxy.JSONCODE{
		Lang:     storage.Lang,
		Filename: "file",
	})
	if err != nil {
		log.Println("Cannot ConfigRunCodeProxy")
		return err
	}

	return nil
}

func (confService *ConfigService) ConfigPluginNetworkStatus(storage model_proxy.JSONDocker) (string, error) {
	var b bytes.Buffer
	if err := utils.Execute(&b,
		exec.Command("docker", "network", "inspect", "bridge"),
		exec.Command("grep", "-A", "5", storage.Server),
		exec.Command("grep", "IPv4Address"),
		// exec.Command("sort", "-r"),
	); err != nil {
		return "", err
	}
	if len(b.String()) < 30 {
		return "", nil
	}
	return b.String()[32:42], nil
}

func (confService *ConfigService) ConfigStopPluginProcess(process_name string) error {
	var processID string
	out, err := utils.RunDockerExec(confService.proxyConfig.DockerID, "", []string{"ps", "-A"})
	if err != nil {
		return err
	}
	words := strings.Fields(out)
	// Get data from process_name generaly is exec-filename
	for i := range words {
		if words[i] == process_name {
			processID = words[i-3]
		}
	}
	_, err = utils.RunDockerExec(confService.proxyConfig.DockerID, "", []string{"kill", "-9", processID})
	if err != nil {
		return err
	}
	confService.ConfigSetStopProxy()
	return nil
}

// Start process in Docker
func (confService *ConfigService) ConfigStartPluginProcess(file string) error {
	_, err := utils.RunDockerExec(confService.proxyConfig.DockerID, "", []string{"/app/go_server/exec-module/" + file, "&"})
	// Start Plugin
	confService.ConfigSetStartProxy()
	return err
}

// GetImage Status
func (confService *ConfigService) startDockerImage() {
	// confService.proxyConfig.EnablePlugin = false
	_, err := confService.configImageInfo()
	if err != nil {
		err = confService.createDockerImage()
		if err != nil {
			log.Fatalln(err)
		}

	}
	confService.proxyConfig.DockerID, err = confService.configImageInfo()
	return
}

func (confService *ConfigService) createDockerImage() error {
	cmdDockerCopy := exec.Command("docker", "run", "-p", "5555:5555", "-p", "5005:5005", "-d", "-t", "--name", utils.DOCKERIMAGE_NAME, "--rm", utils.DOCKERIMAGE_REMOTE_NAME)
	_, err := cmdDockerCopy.Output()
	if err != nil {
		log.Println("Cannot Create DockerImage")
		return err
	}
	return nil
}

// GetImage Status
func (confService *ConfigService) configImageInfo() (string, error) {
	// confService.proxyConfig.EnablePlugin = false
	var b bytes.Buffer
	if err := utils.Execute(&b,
		exec.Command("docker", "ps"),
		exec.Command("grep", utils.DOCKERIMAGE_NAME),
	); err != nil {
		return "", err
	}
	if len(b.String()) < 12 {
		return "", errors.New("No DockerImage start")
	}
	return b.String()[:12], nil
}

// // Ping to Plugin
// func (confService *ConfigService) ConfigCallProxy() error {
// 	return confService.onMessagePortPlugin.Wait()
// }
