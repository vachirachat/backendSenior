package route

import (
	"backendSenior/controller/middleware/auth"
	file_payload "backendSenior/domain/payload/file"
	"backendSenior/domain/service"
	"errors"
	"fmt"
	"github.com/globalsign/mgo"
	"log"

	"github.com/gin-gonic/gin"
	"github.com/globalsign/mgo/bson"

	g "common/utils/ginutils"
)

type FileRouteHandler struct {
	fs   *service.FileService
	room *service.RoomService
	mw   *auth.JWTMiddleware
}

func NewFileRouteHandler(fs *service.FileService, room *service.RoomService, mw *auth.JWTMiddleware) *FileRouteHandler {
	return &FileRouteHandler{
		fs:   fs,
		room: room,
		mw:   mw,
	}
}

func (h *FileRouteHandler) Mount(rg *gin.RouterGroup) {

	rg.POST("/before-upload", h.beforeUploadFile) // todo how make it like POST /room/:roomId
	rg.POST("/room/:roomId/after-upload/:fileId", h.afterUploadFile)
	rg.GET("/room/:roomId/files", h.getRoomFiles)

	rg.GET("/by-id/:fileId", h.getAnyFileDetail) // This shouldn't be used as file messages already contain detail of each file
	rg.POST("/file-url", h.getAnyFileURL)

	rg.POST("/before-upload-image", h.beforeUploadImage)
	rg.POST("/room/:roomId/after-upload-image/:fileId", h.afterUploadImage)
	rg.GET("/room/:roomId/images", h.getRoomImages)

	// POST can be used for actuin
	rg.POST("/delete-file", h.mw.AuthRequired(), g.InjectGin(h.deleteFile))
	rg.POST("/delete-image")
}

func (h *FileRouteHandler) beforeUploadFile(c *gin.Context) {
	id, url, formData, err := h.fs.BeforeUploadFilePOST()
	if err != nil {
		log.Println("before upload file:", err)
		c.JSON(500, gin.H{"status": "error"})
		return
	}

	c.JSON(200, file_payload.BeforeUploadFileResponse{
		URL:      url,
		FileID:   id,
		FormData: formData,
	})
}

func (h *FileRouteHandler) afterUploadFile(c *gin.Context) {
	roomID, ok1 := mustGetObjectID(c, "roomId")
	fileID, ok2 := mustGetObjectID(c, "fileId")
	if !ok1 || !ok2 {
		return
	}
	var uploadMeta service.UploadFileMeta
	err := c.ShouldBindJSON(&uploadMeta)
	if err != nil {
		c.JSON(400, gin.H{"status": "error", "message": "bad request"})
		return
	}
	uploadMeta.RoomID = bson.ObjectIdHex(roomID)

	err = h.fs.AfterUploadFile(fileID, uploadMeta)
	if err != nil {
		log.Println("after upload file:", err)
		c.JSON(500, gin.H{"status": "error", "message": err})
		return
	}

	c.JSON(200, gin.H{"status": "success"})
}

func (h *FileRouteHandler) getRoomFiles(c *gin.Context) {
	roomID, ok := mustGetObjectID(c, "roomId")
	if !ok {
		return
	}
	metas, err := h.fs.GetRoomFileMetas(bson.ObjectIdHex(roomID))
	if err != nil {
		c.JSON(500, gin.H{"status": "error"})
		return
	}

	c.JSON(200, metas)
}

func (h *FileRouteHandler) getAnyFileDetail(c *gin.Context) {
	fileID, ok := mustGetObjectID(c, "fileId")
	if !ok {
		return
	}
	meta, err := h.fs.GetAnyFileMeta(bson.ObjectIdHex(fileID))
	if err != nil {
		log.Printf("get file detail %s: %v", fileID, err)
		c.JSON(500, gin.H{"stauts": "error"})
		return
	}

	c.JSON(200, meta)
	return
}

type fileQuery struct {
	Type   string        `json:"type"`
	FileID bson.ObjectId `json:"id"`
}

func (h *FileRouteHandler) getAnyFileURL(c *gin.Context) {
	// this is POST
	var q fileQuery
	err := c.ShouldBindJSON(&q)
	if err != nil {
		c.JSON(400, gin.H{"status": "error", "message": err})
		return
	}

	url, err := h.fs.GetAnyFileURL(q.Type, q.FileID.Hex())
	if err != nil {
		log.Println("get any file url:", err)
		c.JSON(500, gin.H{"status": "error", "message": err})
		return
	}

	c.JSON(200, gin.H{
		"url": url,
	})
}

// Image
func (h *FileRouteHandler) beforeUploadImage(c *gin.Context) {
	res, err := h.fs.BeforeUploadImagePOST()
	if err != nil {
		log.Println("before upload file:", err)
		c.JSON(500, gin.H{"status": "error"})
		return
	}

	c.JSON(200, res)
}

func (h *FileRouteHandler) afterUploadImage(c *gin.Context) {
	roomID, ok1 := mustGetObjectID(c, "roomId")
	imageFileID, ok2 := mustGetObjectID(c, "fileId")
	if !ok1 || !ok2 {
		return
	}
	var uploadMeta service.UploadImageMeta
	err := c.ShouldBindJSON(&uploadMeta)
	if err != nil {
		c.JSON(400, gin.H{"status": "error", "message": "bad request"})
		return
	}
	uploadMeta.RoomID = bson.ObjectIdHex(roomID)

	err = h.fs.AfterUploadImage(imageFileID, uploadMeta)
	if err != nil {
		log.Println("after upload image:", err)
		c.JSON(500, gin.H{"status": "error", "message": err})
		return
	}

	c.JSON(200, gin.H{"status": "success"})
}

func (h *FileRouteHandler) getRoomImages(c *gin.Context) {
	roomID, ok := mustGetObjectID(c, "roomId")
	if !ok {
		return
	}
	metas, err := h.fs.GetRoomImageMetas(bson.ObjectIdHex(roomID))
	if err != nil {
		c.JSON(500, gin.H{"status": "error"})
		return
	}

	c.JSON(200, metas)
}

func (h *FileRouteHandler) deleteFile(c *gin.Context, input struct {
	// define body here
	Body struct {
		FileID bson.ObjectId `validate:"required"`
	}
}) error {
	b := input.Body

	meta, err := h.fs.GetAnyFileMeta(b.FileID)
	if err != nil {
		if errors.Is(err, mgo.ErrNotFound) {
			return g.NewError(404, err)
		}
		return g.NewError(500, err)
	}

	userID := c.GetString(auth.UserIdField)
	if rooms, err := h.room.GetUserRooms(userID); err != nil {
		return g.NewError(500, fmt.Errorf("error checking user in room: %s", err))
	} else {
		found := false
		for _, roomID := range rooms {
			if roomID.RoomID == meta.RoomID {
				found = true
			}
		}

		if !found {
			return g.NewError(403, errors.New("Forbidden"))
		} else {
			if err := h.fs.DeleteFile(b.FileID); err != nil {
				return g.NewError(500, fmt.Errorf("error deleting file: %s", err))
			} else {
				c.JSON(200, g.Response{true, "deleted file", nil})
				return nil
			}
		}
	}

}
