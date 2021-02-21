package route

import (
	"backendSenior/domain/model"
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/globalsign/mgo/bson"
	"log"
	"math/rand"
	"proxySenior/controller/middleware"
	"proxySenior/domain/encryption"
	model_proxy "proxySenior/domain/model"
	"proxySenior/domain/service"
	"proxySenior/domain/service/key_service"
	"time"
)

type FileRouteHandler struct {
	fs           *service.FileService
	authMw       *middleware.AuthMiddleware
	key          *key_service.KeyService
	upstreamChat *service.ChatUpstreamService
}

func NewFileRouteHandler(fs *service.FileService, authMw *middleware.AuthMiddleware, key *key_service.KeyService, upstreamChat *service.ChatUpstreamService) *FileRouteHandler {
	return &FileRouteHandler{
		fs:           fs,
		authMw:       authMw,
		key:          key,
		upstreamChat: upstreamChat,
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

	meta, err := h.fs.GetAnyFileMeta(fileID)
	if err != nil {
		log.Println("get file: get meta:", err)
		c.JSON(500, gin.H{"status": "error", "message": "error"})
		return
	}

	fileData, err := h.fs.GetAnyFile(fileType, fileID)
	if err != nil {
		log.Println("get file: service:", err)
		c.JSON(500, gin.H{"status": "error", "message": "error"})
		return
	}

	log.Printf("[v] file meta %+v\n", meta)
	keys, err := h.getKeyFromRoom(meta.RoomID.Hex())
	if err != nil {
		log.Println("get file: get key for decrypt:", err)
		c.JSON(500, gin.H{"status": "error", "message": "error"})
		return
	}
	key := keyFor(keys, meta.CreatedAt)
	fileData, err = encryption.AESDecrypt(fileData, key)
	if err != nil {
		log.Println("get file: decrypt error:", err)
		c.JSON(500, gin.H{"status": "error", "message": "error"})
		return
	}

	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s", meta.FileName))
	c.Header("Content-Length", fmt.Sprint(len(fileData)))
	c.Data(200, "application/octet-stream", fileData)

}

func (h *FileRouteHandler) uploadFile(c *gin.Context) {
	userID := c.GetString(middleware.UserIdField)
	roomID := c.Param("roomId")
	if !bson.IsObjectIdHex(roomID) {
		c.JSON(400, gin.H{"status": "error", "message": "bad roomId param"})
		return
	}

	fmt.Println(c.Request.Header)
	fmt.Println(c.Request.Form)

	header, err := c.FormFile("file")
	if err != nil {
		log.Println("get file in form", err)
		c.JSON(500, err)
		return
	}

	tempPath := fmt.Sprintf("/tmp/upload_%d", rand.Int31())

	err = c.SaveUploadedFile(header, tempPath)
	if err != nil {
		c.JSON(500, gin.H{"status": "couldn't save uploaded file"})
		return
	}
	now := time.Now()
	keys, err := h.getKeyFromRoom(roomID)
	if err != nil {
		log.Println("get key for room", err)
		c.JSON(500, fmt.Errorf("get key for room: %w", err))
		return
	}
	key := keyFor(keys, now)

	fileID, metaAsync, err := h.fs.UploadFile(roomID, userID, key, model.FileDetail{
		Path:        tempPath,
		Name:        header.Filename,
		Size:        int(header.Size),
		CreatedTime: now,
	})

	//fileData, err = encryption.AESEncrypt(fileData, key)
	//if err != nil {
	//	log.Println("encrypt file", err)
	//	c.JSON(500, fmt.Errorf("encrypt file: %w", err))
	//	return
	//}
	//fileID, err := h.fs.UploadFile(roomID, userID, header.Filename, now, bytes.NewReader(fileData))
	//if err != nil {
	//	log.Println("upload file", err)
	//	c.JSON(500, fmt.Errorf("upload file: %w", err))
	//	return
	//}

	// TOOD: remove these

	// send message to room
	go func() {
		meta := <-metaAsync
		if meta.FileID == "" {
			log.Println("[upload file handler] task failed")
			return
		}
		metaBytes, err := json.Marshal(meta)
		if err != nil {
			fmt.Printf("error marshal: %s\n", err)
			return
		}

		metaBytes, err = encryption.AESEncrypt(metaBytes, key)
		if err != nil {
			log.Println("upload file: encrypt meta: %w\n", err)
			c.JSON(500, err)
			return
		}

		metaBytes = encryption.B64Encode(metaBytes)
		h.upstreamChat.SendMessage(model.Message{
			TimeStamp: now,
			RoomID:    bson.ObjectIdHex(roomID),
			UserID:    bson.ObjectIdHex(userID),
			ClientUID: "foo",             // TODO: this isn't needed?
			Data:      string(metaBytes), // tell client the meta
			Type:      "FILE",
		})
	}()

	c.JSON(202, gin.H{
		"fileId": fileID,
	})
}

// TODO: duplicated code
//getKeyFromRoom determine where to get key and get the key
func (h *FileRouteHandler) getKeyFromRoom(roomID string) ([]model_proxy.KeyRecord, error) {
	local, err := h.key.IsLocal(roomID)
	if err != nil {
		return nil, fmt.Errorf("error deftermining locality ok key %v", err)
	}

	var keys []model_proxy.KeyRecord
	if local {
		fmt.Println("[message] use LOCAL key for", roomID)
		keys, err = h.key.GetKeyLocal(roomID)
		if err != nil {
			return nil, fmt.Errorf("error getting key locally %v", err)
		}
	} else {
		fmt.Println("[message] use REMOTE key for room", roomID)
		keys, err = h.key.GetKeyRemote(roomID)
		if err != nil {
			return nil, fmt.Errorf("error getting key remotely %v", err)
		}
	}

	return keys, nil
}
