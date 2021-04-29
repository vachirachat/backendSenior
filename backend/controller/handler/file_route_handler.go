package route

import (
	"backendSenior/controller/middleware/auth"
	"backendSenior/domain/dto"
	file_payload "backendSenior/domain/payload/file"
	"backendSenior/domain/service"
	"backendSenior/utills"
	"errors"
	"fmt"
	"log"

	"github.com/globalsign/mgo"

	"github.com/gin-gonic/gin"
	"github.com/globalsign/mgo/bson"

	g "common/utils/ginutils"
)

type FileRouteHandler struct {
	fs       *service.FileService
	room     *service.RoomService
	mw       *auth.JWTMiddleware
	validate *utills.StructValidator
}

func NewFileRouteHandler(fs *service.FileService, room *service.RoomService, mw *auth.JWTMiddleware, validate *utills.StructValidator) *FileRouteHandler {
	return &FileRouteHandler{
		fs:       fs,
		room:     room,
		mw:       mw,
		validate: validate,
	}
}

func (h *FileRouteHandler) Mount(rg *gin.RouterGroup) {

	rg.POST("/before-upload", g.InjectGin(h.beforeUploadFile)) // todo how make it like POST /room/:roomId
	rg.POST("/room/:roomId/after-upload/:fileId", g.InjectGin(h.afterUploadFile))
	rg.GET("/room/:roomId/files", g.InjectGin(h.getRoomFiles))

	rg.GET("/by-id/:fileId", g.InjectGin(h.getAnyFileDetail)) // This shouldn't be used as file messages already contain detail of each file
	rg.POST("/file-url", g.InjectGin(h.getAnyFileURL))

	rg.POST("/before-upload-image", g.InjectGin(h.beforeUploadImage))
	rg.POST("/room/:roomId/after-upload-image/:fileId", g.InjectGin(h.afterUploadImage))
	rg.GET("/room/:roomId/images", g.InjectGin(h.getRoomImages))

	// POST can be used for actuin
	rg.POST("/delete-file", h.mw.AuthRequired("user", "edit"), g.InjectGin(h.deleteFile))
	rg.POST("/delete-image", h.mw.AuthRequired("user", "edit"), g.InjectGin(h.deleteImage))
}

func (h *FileRouteHandler) beforeUploadFile(c *gin.Context, req struct{}) error {
	id, url, formData, err := h.fs.BeforeUploadFilePOST()
	if err != nil {
		log.Println("before upload file:", err)
		// c.JSON(500, gin.H{"status": "error"})
		return err
	}

	c.JSON(200, file_payload.BeforeUploadFileResponse{
		URL:      url,
		FileID:   id,
		FormData: formData,
	})
	return nil
}

func (h *FileRouteHandler) afterUploadFile(c *gin.Context, req struct{ Body dto.UploadFileMeta }) error {
	roomID, ok1 := mustGetObjectID(c, "roomId")
	fileID, ok2 := mustGetObjectID(c, "fileId")
	if !ok1 || !ok2 {
		return g.NewError(400, "invalid fileID or roomID")
	}
	// var uploadMeta dto.UploadFileMeta
	// err := c.ShouldBindJSON(&uploadMeta)
	// if err != nil {
	// 	c.JSON(400, gin.H{"status": "error", "message": "bad request"})
	// 	return err
	// }
	b := req.Body
	err := h.validate.ValidateStruct(b)
	if err != nil {
		return g.NewError(400, "bad Body UploadFileMeta")
	}
	b.RoomID = bson.ObjectIdHex(roomID)

	err = h.fs.AfterUploadFile(fileID, b)
	if err != nil {
		// log.Println("after upload file:", err)
		// c.JSON(500, gin.H{"status": "error", "message": err})
		return g.NewError(500, "error after upload file")
	}

	c.JSON(200, gin.H{"status": "success"})
	return nil
}

func (h *FileRouteHandler) getRoomFiles(c *gin.Context, req struct{}) error {
	roomID, ok := mustGetObjectID(c, "roomId")
	if !ok {
		return nil
	}
	metas, err := h.fs.GetRoomFileMetas(bson.ObjectIdHex(roomID))
	if err != nil {
		// c.JSON(500, gin.H{"status": "error"})
		return err
	}

	c.JSON(200, metas)
	return nil
}

func (h *FileRouteHandler) getAnyFileDetail(c *gin.Context, req struct{}) error {
	fileID, ok := mustGetObjectID(c, "fileId")
	if !ok {
		return nil
	}
	meta, err := h.fs.GetAnyFileMeta(bson.ObjectIdHex(fileID))
	if err != nil {
		log.Printf("get file detail %s: %v", fileID, err)
		// c.JSON(500, gin.H{"stauts": "error"})
		return err
	}

	c.JSON(200, meta)
	return nil
}

func (h *FileRouteHandler) getAnyFileURL(c *gin.Context, req struct{ Body dto.FileQuery }) error {
	// this is POST
	// var q fileQuery
	// err := c.ShouldBindJSON(&q)
	// if err != nil {
	// 	c.JSON(400, gin.H{"status": "error", "message": err})
	// 	return

	// }
	b := req.Body
	err := h.validate.ValidateStruct(b)
	if err != nil {
		return g.NewError(400, "bad Body FileQuery")
	}
	if !bson.IsObjectIdHex(b.FileID.Hex()) {
		return g.NewError(400, "invalid FileID in path")
	}

	url, err := h.fs.GetAnyFileURL(b.Type, b.FileID.Hex())
	if err != nil {
		log.Println("get any file url:", err)
		// c.JSON(500, gin.H{"status": "error", "message": err})
		return err
	}

	c.JSON(200, gin.H{
		"url": url,
	})
	return nil
}

// Image
func (h *FileRouteHandler) beforeUploadImage(c *gin.Context, req struct{}) error {
	res, err := h.fs.BeforeUploadImagePOST()
	if err != nil {
		log.Println("before upload file:", err)
		// c.JSON(500, gin.H{"status": "error"})
		return err
	}

	c.JSON(200, res)
	return nil
}

func (h *FileRouteHandler) afterUploadImage(c *gin.Context, req struct{ Body dto.UploadImageMeta }) error {
	roomID, ok1 := mustGetObjectID(c, "roomId")
	imageFileID, ok2 := mustGetObjectID(c, "fileId")
	if !ok1 || !ok2 {
		return g.NewError(400, "invalid fileID or roomID")
	}

	// var uploadMeta dto.UploadImageMeta
	// err := c.ShouldBindJSON(&uploadMeta)
	// if err != nil {
	// 	c.JSON(400, gin.H{"status": "error", "message": "bad request"})
	// 	return err
	// }
	b := req.Body
	err := h.validate.ValidateStruct(b)
	if err != nil {
		return g.NewError(400, "bad Body UploadImageMeta")
	}
	b.RoomID = bson.ObjectIdHex(roomID)

	err = h.fs.AfterUploadImage(imageFileID, b)
	if err != nil {
		log.Println("after upload image:", err)
		// c.JSON(500, gin.H{"status": "error", "message": err})
		return err
	}

	c.JSON(200, gin.H{"status": "success"})
	return nil
}

func (h *FileRouteHandler) getRoomImages(c *gin.Context, req struct{}) error {
	roomID, ok := mustGetObjectID(c, "roomId")
	if !ok {
		return g.NewError(400, "invalid roomID")
	}
	metas, err := h.fs.GetRoomImageMetas(bson.ObjectIdHex(roomID))
	if err != nil {
		// c.JSON(500, gin.H{"status": "error"})
		return err
	}

	c.JSON(200, metas)
	return nil
}

func (h *FileRouteHandler) deleteFile(c *gin.Context, input struct {
	// define body here
	Body struct {
		FileID bson.ObjectId `validate:"required"`
	}
}) error {
	b := input.Body
	err := h.validate.ValidateStruct(b)
	if err != nil {
		return g.NewError(400, "bad Body FileID")
	}

	meta, err := h.fs.GetAnyFileMeta(b.FileID)
	if err != nil {
		if errors.Is(err, mgo.ErrNotFound) {
			return g.NewError(404, err.Error())
		}
		return g.NewError(500, err.Error())
	}

	userID := c.GetString(auth.UserIdField)
	if rooms, err := h.room.GetUserRooms(userID); err != nil {
		return g.NewError(500, fmt.Sprintf("error checking user in room: %s", err))
	} else {
		found := false
		for _, roomID := range rooms {
			if roomID.RoomID == meta.RoomID {
				found = true
			}
		}

		if !found {
			return g.NewError(403, "you are not in the room")
		} else {
			if meta.UserID.Hex() != userID {
				return g.NewError(403, "forbidden: not your file")
			}

			if err := h.fs.DeleteFile(b.FileID); err != nil {
				return g.NewError(500, fmt.Sprintf("error deleting file: %s", err))
			} else {
				c.JSON(200, g.Response{true, "deleted file", nil})
				return nil
			}
		}
	}

}

func (h *FileRouteHandler) deleteImage(c *gin.Context, input struct {
	// define body here
	Body struct {
		FileID bson.ObjectId `validate:"required"`
	}
}) error {
	b := input.Body
	err := h.validate.ValidateStruct(b)
	if err != nil {
		return g.NewError(400, "bad Body FileID")
	}

	meta, err := h.fs.GetAnyFileMeta(b.FileID)
	if err != nil {
		if errors.Is(err, mgo.ErrNotFound) {
			return g.NewError(404, err.Error())
		}
		return g.NewError(500, err.Error())
	}

	userID := c.GetString(auth.UserIdField)
	if rooms, err := h.room.GetUserRooms(userID); err != nil {
		return g.NewError(500, fmt.Sprintf("error checking user in room: %s", err))
	} else {
		found := false
		for _, roomID := range rooms {
			if roomID.RoomID == meta.RoomID {
				found = true
			}
		}

		if !found {
			return g.NewError(403, "forbidden: not in room")
		} else {
			if meta.UserID.Hex() != userID {
				return g.NewError(403, "forbidden: not your image")
			}

			if err := h.fs.DeleteImage(b.FileID); err != nil {
				return g.NewError(500, fmt.Sprintf("error deleting image: %s", err))
			} else {
				c.JSON(200, g.Response{true, "deleted image", nil})
				return nil
			}
		}
	}

}
