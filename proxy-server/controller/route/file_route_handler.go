package route

import (
	"io"
	"log"
	"proxySenior/controller/middleware"
	"proxySenior/domain/service"

	"github.com/gin-gonic/gin"
	"github.com/globalsign/mgo/bson"
)

type FileRouteHandler struct {
	fs     *service.FileService
	authMw *middleware.AuthMiddleware
}

func NewFileRouteHandler(fs *service.FileService, authMw *middleware.AuthMiddleware) *FileRouteHandler {
	return &FileRouteHandler{
		fs:     fs,
		authMw: authMw,
	}
}

func (h *FileRouteHandler) Mount(rg *gin.RouterGroup) {

	rg.GET("/file/:fileId", h.getFile)
	rg.POST("/room/:roomId", h.authMw.AuthRequired(), h.uploadFile)

}

func (h *FileRouteHandler) getFile(c *gin.Context) {
	fileType := c.DefaultQuery("type", "file")
	if fileType != "image" && fileType != "file" {
		c.JSON(400, gin.H{"status": "error", "message": "bad file type (image/file only)"})
		return
	}

	fileID := c.Param("fileId")
	if !bson.IsObjectIdHex(fileID) {
		c.JSON(400, gin.H{"status": "error", "message": "bad fileId param"})
		return
	}

	file, err := h.fs.GetAnyFile(fileType, fileID)
	if err != nil {
		log.Println("get file: service:", err)
		c.JSON(500, gin.H{"status": "error", "message": "error"})
		return
	}

	c.Header("Content-Type", "application/octet-stream")
	c.Status(200)
	buf := make([]byte, 2<<10) // 2KB ?
	for {
		n, err := file.Read(buf)
		c.Writer.Write(buf[:n])
		if err != nil {
			if err == io.EOF {
				break
			} else {
				c.Status(500)
				return
			}
		}
	}
}

func (h *FileRouteHandler) uploadFile(c *gin.Context) {
	userID := c.GetString(middleware.UserIdField)
	roomID := c.Param("roomId")
	if !bson.IsObjectIdHex(roomID) {
		c.JSON(400, gin.H{"status": "error", "message": "bad roomId param"})
		return
	}

	header, err := c.FormFile("file")
	if err != nil {
		log.Println("get file", err)
		c.JSON(500, err)
		return
	}
	file, err := header.Open()
	if err != nil {
		log.Println("open file", err)
		c.JSON(500, err)
		return
	}

	err = h.fs.UploadFile(bson.ObjectIdHex(roomID), bson.ObjectIdHex(userID), header.Filename, file)
	if err != nil {
		log.Println("upload file", err)
		c.JSON(500, err)
		return
	}

	c.JSON(200, "OK")
}
