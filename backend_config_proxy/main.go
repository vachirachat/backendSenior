package main

import (
	"bufio"
	"bytes"
	"net/http"
	"os/exec"
	"strings"

	"fmt"
	"io"
	"log"
	"os"

	docker "github.com/fsouza/go-dockerclient"
	"github.com/gin-gonic/gin"
)

var containerID string
var localFilePath string
var localPlainFilePath string
var dockerFilePath string
var PORT string

func main() {
	envPath, err := os.Getwd()
	if err != nil {
		log.Println(err)
	}
	containerID = ""
	if containerID == "" {
		log.Panicln("Build/Run Dockerfile with in File ", envPath, "docker to get Dockerid")
	}
	localFilePath = envPath + "/file_upload/"
	localPlainFilePath = envPath + "/golang-code/"
	dockerFilePath = "/app/go_server/"
	PORT = "localhost:7070"
	r := setupRoutes()
	r.Run(PORT)

}

func runDockerExec(containerID string, cmdInput []string) string {

	dockerClient, err := docker.NewClientFromEnv()
	if err != nil {
		log.Panicln("Error: %s", err)
	}
	dockerExOpt := docker.CreateExecOptions{
		AttachStderr: true,
		AttachStdin:  true,
		AttachStdout: true,
		Tty:          false,
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
	router.POST("/upload", uploadFile)
	router.POST("/compilecode", compileCode)
	router.GET("/run", runFile)
	router.GET("/kill", killProcess)
	router.LoadHTMLGlob("templates/*")
	router.GET("/", func(c *gin.Context) {
		c.HTML(
			http.StatusOK,
			"fileUp.html",
			gin.H{
				"title": "Web",
			},
		)
	})

	return router
}
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

	cmdCovert := exec.Command("docker", "cp", localFilePath, containerID+":"+dockerFilePath)
	log.Println("Run command  >>>>", "docker", "cp", localFilePath+handler.Filename, containerID+":"+dockerFilePath)
	_, err = cmdCovert.Output()
	if err != nil {
		log.Println("Cannot cmdCovert")
		log.Fatal(err)
	}

	cmdRemove := exec.Command("rm", localFilePath+handler.Filename)
	_, err = cmdRemove.Output()
	if err != nil {
		log.Println("Cannot remove file-type")
		log.Fatal(err)
	}
	fmt.Fprintf(context.Writer, "Successfully Uploaded File\n")

	return

}

func killProcess(context *gin.Context) {
	var processID string
	process_name, ok := context.Request.URL.Query()["process_name"]
	log.Println(process_name[0])
	if context.Request.Method == "GET" || !ok {
		out := runDockerExec(containerID, []string{"ps", "-A"})
		words := strings.Fields(out)

		for i := range words {
			if words[i] == process_name[0] {
				processID = words[i-3]
			}
		}
		_ = runDockerExec(containerID, []string{"kill", "-9", processID})
	}

}

func runFile(context *gin.Context) {
	file, ok := context.Request.URL.Query()["file"]
	log.Println(file[0])
	if context.Request.Method == "GET" || !ok {
		_ = runDockerExec(containerID, []string{"/app/go_server/file_upload/" + file[0], "&"})
	}

}

type JSONCODE struct {
	Code string `json:"code"`
	Lang string `json:"lang"`
}

func compileCode(context *gin.Context) {
	var storage JSONCODE
	err := context.ShouldBindJSON(&storage)
	if err != nil {
		context.JSON(http.StatusBadRequest, gin.H{"status": "error"})
		return
	}
	log.Println(storage)
	f, err := os.Create(localPlainFilePath + "test-txt.go")
	defer f.Close()
	w := bufio.NewWriter(f)
	n4, err := w.WriteString(storage.Code)
	fmt.Printf("wrote %d bytes\n", n4)
	w.Flush()

	// Copy to docker
	cmdCovert := exec.Command("docker", "cp", localPlainFilePath, containerID+":"+dockerFilePath)
	_, err = cmdCovert.Output()
	if err != nil {
		log.Println("Cannot cmdCovert")
		log.Fatal(err)
	}

	cmdRemove := exec.Command("rm", localPlainFilePath+"test-txt.go")
	_, err = cmdRemove.Output()
	if err != nil {
		log.Println("Cannot remove file-type")
		log.Fatal(err)
	}
	fmt.Fprintf(context.Writer, "Successfully Uploaded File\n")

}

// docker cp /Users/waritphon/code-fast-test/golang-upload-exec/file_upload/docker-cheatsheet e62d0337d995:/app/go_server/file_upload/
