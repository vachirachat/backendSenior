package route

import (
	"backendSenior/domain/model"
	g "common/utils/ginutils"
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/globalsign/mgo/bson"
	"github.com/go-resty/resty/v2"
	"log"
	"math/rand"
	"net/url"
	"proxySenior/config"
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
	rg.POST("/room/:roomId/files", h.authMw.AuthRequired(), h.uploadFile)
	rg.POST("/room/:roomId/images", h.authMw.AuthRequired(), h.uploadImage)

	rg.GET("/room/:id/files", h.authMw.AuthRequired(), g.InjectGin(h.listFiles))
	rg.GET("/room/:id/images", h.authMw.AuthRequired(), g.InjectGin(h.listImages))
	//rg.GET("/room/:roomID/images", h.authMw.AuthRequired(), h.getImages)
}

func (h *FileRouteHandler) getFile(c *gin.Context) {
	fileType := c.DefaultQuery("type", "file")
	if fileType != "image" && fileType != "file" {
		c.JSON(400, gin.H{"status": "error", "message": "bad file type (image/file only)"})
		return
	}

	fileID := c.Param("fileId")                // ID for lookup meta
	overrideID := c.DefaultQuery("id", fileID) // ID for actual file, it'll be different if it's thumbnail
	if !bson.IsObjectIdHex(fileID) || !bson.IsObjectIdHex(overrideID) {
		c.JSON(400, gin.H{"status": "error", "message": "bad fileId param"})
		return
	}

	// get meta for fileID
	meta, err := h.fs.GetAnyFileMeta(fileID)
	if err != nil {
		log.Println("get file: get meta:", err)
		c.JSON(500, gin.H{"status": "error", "message": "error"})
		return
	}

	// get data for overrideID
	fileData, err := h.fs.GetAnyFile(fileType, overrideID)
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
			FileID:    meta.FileID,
			RoomID:    bson.ObjectIdHex(roomID),
			UserID:    bson.ObjectIdHex(userID),
			ClientUID: "foo",             // TODO: this isn't needed?
			Data:      string(metaBytes), // tell client the meta
			Type:      model.MsgFile,
		})
	}()

	c.JSON(202, gin.H{
		"fileId": fileID,
	})
}

func (h *FileRouteHandler) uploadImage(c *gin.Context) {
	userID := c.GetString(middleware.UserIdField)
	roomID := c.Param("roomId")
	if !bson.IsObjectIdHex(roomID) {
		c.JSON(400, gin.H{"status": "error", "message": "bad roomId param"})
		return
	}

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

	imageID, metaAsync, err := h.fs.UploadImage(roomID, userID, key, model.FileDetail{
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
			log.Println("[upload image handler] task failed")
			return
		}
		metaBytes, err := json.Marshal(meta)
		if err != nil {
			fmt.Printf("error marshal: %s\n", err)
			return
		}

		metaBytes, err = encryption.AESEncrypt(metaBytes, key)
		if err != nil {
			log.Println("upload image: encrypt meta: %w\n", err)
			c.JSON(500, err)
			return
		}

		metaBytes = encryption.B64Encode(metaBytes)
		h.upstreamChat.SendMessage(model.Message{
			TimeStamp: now,
			FileID:    meta.FileID,
			RoomID:    bson.ObjectIdHex(roomID),
			UserID:    bson.ObjectIdHex(userID),
			ClientUID: "foo",             // TODO: this isn't needed?
			Data:      string(metaBytes), // tell client the meta
			Type:      model.MsgImage,
		})
	}()

	c.JSON(202, gin.H{
		"fileId": imageID,
	})
}

func (h *FileRouteHandler) listFiles(c *gin.Context, req struct{}) error {
	id := c.Param("id")
	if !bson.IsObjectIdHex(id) {
		return g.NewError(400, "bad room Id")
	}

	endPoint := url.URL{
		Scheme: "http",
		Host:   config.ControllerOrigin,
		Path:   fmt.Sprintf("/api/v1/file/room/%s/files", id),
	}

	clnt := resty.New()
	keys, err := h.getKeyFromRoom(id)
	if err != nil {
		log.Printf("err getting key to decrypt: %s", err)
		return fmt.Errorf("err getting key to decrypt: %s", err)
	}

	var res []model.FileMeta
	if _, err := clnt.R().SetResult(&res).Get(endPoint.String()); err != nil {
		log.Printf("can't get file list from controller: %s", err)
		return fmt.Errorf("can't get file list from controller: %s", err)

	}
	for i := range res {
		date := res[i].CreatedAt
		key := keyFor(keys, date)
		if key == nil {
			res[i].FileName = "bad_file_timestamp"
			continue
		}
		encryptedFileName, err := encryption.B64Decode([]byte(res[i].FileName))
		if err != nil {
			res[i].FileName = "bad_base64_filename"
			continue
		}
		fileName, err := encryption.AESDecrypt(encryptedFileName, key)
		if err != nil {
			res[i].FileName = "error_decrypting_filename"
		}
		res[i].FileName = string(fileName)
	}

	c.JSON(200, res)
	return nil
}

func (h *FileRouteHandler) listImages(c *gin.Context, req struct{}) error {
	id := c.Param("id")
	if !bson.IsObjectIdHex(id) {
		return g.NewError(400, "bad room Id")
	}

	endPoint := url.URL{
		Scheme: "http",
		Host:   config.ControllerOrigin,
		Path:   fmt.Sprintf("/api/v1/file/room/%s/images", id),
	}

	clnt := resty.New()
	keys, err := h.getKeyFromRoom(id)
	if err != nil {
		log.Printf("err getting key to decrypt: %s", err)
		return fmt.Errorf("err getting key to decrypt: %s", err)
	}

	var res []model.FileMeta
	if _, err := clnt.R().SetResult(&res).Get(endPoint.String()); err != nil {
		log.Printf("can't get file list from controller: %s", err)
		return fmt.Errorf("can't get file list from controller: %s", err)

	}
	for i := range res {
		date := res[i].CreatedAt
		key := keyFor(keys, date)
		if key == nil {
			res[i].FileName = "bad_file_timestamp"
			continue
		}
		encryptedFileName, err := encryption.B64Decode([]byte(res[i].FileName))
		if err != nil {
			res[i].FileName = "bad_base64_filename"
			continue
		}
		fileName, err := encryption.AESDecrypt(encryptedFileName, key)
		if err != nil {
			res[i].FileName = "error_decrypting_filename"
		}
		res[i].FileName = string(fileName)
	}

	c.JSON(200, res)
	return nil
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
		//fmt.Println("[message] use LOCAL key for", roomID)
		keys, err = h.key.GetKeyLocal(roomID)
		if err != nil {
			return nil, fmt.Errorf("error getting key locally %v", err)
		}
	} else {
		//fmt.Println("[message] use REMOTE key for room", roomID)
		keys, err = h.key.GetKeyRemote(roomID)
		if err != nil {
			return nil, fmt.Errorf("error getting key remotely %v", err)
		}
	}

	return keys, nil
}
