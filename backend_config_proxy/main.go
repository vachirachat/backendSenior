package main

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"strings"

	PipeService "backend_config_proxy/service"
	"log"

	docker "github.com/fsouza/go-dockerclient"
	"github.com/gin-gonic/gin"
)

var containerID string
var localFilePath string
var localPlainFilePath string
var dockerFilePath string
var PORT string

func main() {
	containerID = "8108a7bfb5a0"
	absPath, err := os.Getwd()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	fmt.Println(absPath)

	localFilePath = absPath + "/file_upload/"
	localPlainFilePath = absPath + "/temporary-code"
	dockerFilePath = "/app/go_server"
	PORT = "localhost:7070"
	r := setupRoutes()
	r.Run(PORT)

}

func runDockerExec(containerID string, workingDir string, cmdInput []string) string {
	// dockerAPIversion, _ := docker.NewAPIVersion()
	// dockerClient, err := docker.NewClientFromEnv()
	apiVersionString := "1.41"
	dockerClient, err := docker.NewVersionedClientFromEnv(apiVersionString)
	if err != nil {
		log.Panicln(">> ", err)
		return ""
	}
	dockerClient.SkipServerVersionCheck = apiVersionString == ""

	if err != nil {
		log.Panicln("Error: %s", err)
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
		log.Panicln("Error: %s", err)
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
		log.Panicln("Error: %s", err)
	}
	// log.Println("stdout: %+s", stdout.String())
	return stdout.String()

}

func setupRoutes() *gin.Engine {
	router := gin.Default()
	router.POST("/config/upload", uploadFile)
	router.POST("/config/compilecode", compileCode)
	router.POST("/config/runcode", runFileCode)
	router.GET("/config/run", runFile)
	router.GET("/config/kill", killProcess)

	router.POST("/docker/getip", dockerIP)

	return router
}

// Exec run
func uploadFile(context *gin.Context) {
	context.Request.ParseMultipartForm(10 << 20)
	file, handler, err := context.Request.FormFile("myFile")
	if err != nil {
		fmt.Println("Error Retrieving the File")
		fmt.Println(err)
		return
	}
	defer file.Close()
	fmt.Printf("Uploaded File: %+v\n", handler.Filename)
	fmt.Printf("File Size: %+v\n", handler.Size)
	fmt.Printf("MIME Header: %+v\n", handler.Header)

	// Create file
	dst, err := os.Create("./file_upload/" + handler.Filename)
	defer dst.Close()
	if err != nil {
		http.Error(context.Writer, err.Error(), http.StatusInternalServerError)
		return
	}

	// Copy the uploaded file to the created file on the filesystem
	if _, err := io.Copy(dst, file); err != nil {
		http.Error(context.Writer, err.Error(), http.StatusInternalServerError)
		return
	}
	cmdChmod := exec.Command("chmod", "+x", localFilePath+handler.Filename)
	_, err = cmdChmod.Output()
	if err != nil {
		log.Println("Cannot change file-type")
		log.Fatal(err)
	}

	cmdDockerCopy := exec.Command("docker", "cp", localFilePath+handler.Filename, containerID+":"+dockerFilePath+"/exec-module")
	_, err = cmdDockerCopy.Output()
	if err != nil {
		log.Println("Cannot cmdDockerCopy")
		log.Fatal(err)
	}

	// cmdRemove := exec.Command("rm", localFilePath+handler.Filename)
	// _, err = cmdRemove.Output()
	// if err != nil {
	// 	log.Println("Cannot remove file-type")
	// 	log.Fatal(err)
	// }
	// fmt.Fprintf(context.Writer, "Successfully Uploaded File\n")

	return

}

func killProcess(context *gin.Context) {
	var processID string
	process_name, ok := context.Request.URL.Query()["process_name"]
	log.Println(process_name[0])
	if context.Request.Method == "GET" || !ok {
		out := runDockerExec(containerID, "", []string{"ps", "-A"})
		words := strings.Fields(out)
		// Get data from process_name generaly is exec-filename
		for i := range words {
			if words[i] == process_name[0] {
				processID = words[i-3]
			}
		}
		_ = runDockerExec(containerID, "", []string{"kill", "-9", processID})
	}

}

func runFile(context *gin.Context) {
	file, ok := context.Request.URL.Query()["file"]
	log.Println(file[0])
	if context.Request.Method == "GET" || !ok {
		_ = runDockerExec(containerID, "", []string{"/app/go_server/exec-module/" + file[0], "&"})
	}

}

// Code run

type JSONCODE struct {
	Code     string `json:"Config"`
	Lang     string `json:"lang"`
	Filename string `json:"filename"`
}

func runFileCode(context *gin.Context) {
	var storage JSONCODE
	err := context.ShouldBindJSON(&storage)
	if err != nil {
		context.JSON(http.StatusBadRequest, gin.H{"status": "error"})
		return
	}

	if storage.Lang == "go" {
		out := runDockerExec(containerID, "/app/go_server/"+storage.Lang+"-module/", []string{"go", "run", storage.Filename + "." + storage.Lang})
		log.Println("out file go Run >>>>", out)
	} else if storage.Lang == "py" {
		out := runDockerExec(containerID, "/app/go_server/"+storage.Lang+"-module/", []string{"python3", storage.Filename + "." + storage.Lang})
		log.Println("out file py Run >>>>", out)
		// _ = runDockerExec(containerID, []string{"python3", "run", storage.Lang + "-module/file." + storage.Lang, "&"})
	} else if storage.Lang == "js" {
		out := runDockerExec(containerID, "/app/go_server/"+storage.Lang+"-module/", []string{"node", storage.Filename + "." + storage.Lang})
		log.Println("out file js Run >>>>", out)

	}

}

func compileCode(context *gin.Context) {
	var storage JSONCODE
	err := context.ShouldBindJSON(&storage)
	if err != nil {
		context.JSON(http.StatusBadRequest, gin.H{"status": "error"})
		return
	}
	if storage.Lang == "" {
		context.JSON(http.StatusBadRequest, gin.H{"status": "error"})
		return
	}
	log.Println("Lang", storage.Lang)
	// File Upload
	f, err := os.Create(localPlainFilePath + "/" + storage.Lang + "/file." + storage.Lang)
	defer f.Close()
	w := bufio.NewWriter(f)
	n4, err := w.WriteString(storage.Code)
	fmt.Printf("wrote %d bytes\n", n4)
	w.Flush()

	// Upload to Docker
	log.Println("docker", "cp", localPlainFilePath+"/"+storage.Lang+"/file."+storage.Lang, containerID+":"+dockerFilePath+"/go-module")
	if storage.Lang == "go" {
		// Copy to docker
		cmdCovert := exec.Command("docker", "cp", localPlainFilePath+"/"+storage.Lang+"/file."+storage.Lang, containerID+":"+dockerFilePath+"/go-module")
		_, err = cmdCovert.Output()
		if err != nil {
			log.Println("Cannot cmdCovert")
			log.Fatal(err)
		}
	} else if storage.Lang == "py" {
		// Copy to docker
		cmdCovert := exec.Command("docker", "cp", localPlainFilePath+"/"+storage.Lang+"/file."+storage.Lang, containerID+":"+dockerFilePath+"/py-module")
		_, err = cmdCovert.Output()
		if err != nil {
			log.Println("Cannot cmdCovert")
			log.Fatal(err)
		}
	} else if storage.Lang == "js" {
		cmdCovert := exec.Command("docker", "cp", localPlainFilePath+"/"+storage.Lang+"/file."+storage.Lang, containerID+":"+dockerFilePath+"/js-module")
		_, err = cmdCovert.Output()
		if err != nil {
			log.Println("Cannot cmdCovert")
			log.Fatal(err)
		}
	}

	// cmdRemove := exec.Command("rm", localPlainFilePath+"/"+storage.Lang+"/file."+storage.Lang)
	// _, err = cmdRemove.Output()
	// if err != nil {
	// 	log.Println("Cannot remove file-type")
	// 	log.Fatal(err)
	// }

}

// Docker Manage

type JSONDocker struct {
	Server string `json:"server"`
	Status string `json:"status"`
	IP     string `json:"ip"`
}

func dockerIP(context *gin.Context) {
	var storage JSONDocker
	err := context.ShouldBindJSON(&storage)
	if err != nil {
		log.Println("err -binding")
		context.JSON(http.StatusBadRequest, gin.H{"status": "error"})
		return
	}
	var b bytes.Buffer
	if err := PipeService.Execute(&b,
		exec.Command("docker", "network", "inspect", "bridge"),
		exec.Command("grep", "-A", "5", storage.Server),
		exec.Command("grep", "IPv4Address"),
		// exec.Command("sort", "-r"),
	); err != nil {
		log.Fatalln(err)
	}

	log.Println(string(b.String()))
	context.JSON(http.StatusAccepted, gin.H{"status": b.String()})
}
