package route

import (
	authMw "backendSenior/controller/middleware/auth"
	"backendSenior/domain/dto"
	"backendSenior/domain/service"
	"backendSenior/utills"
	"bytes"
	g "common/utils/ginutils"
	"fmt"
	"image"
	"io"
	"io/ioutil"
	"log"
	"os"

	"github.com/disintegration/imaging"
	"github.com/gin-gonic/gin"
	"github.com/globalsign/mgo/bson"
)

type StickerRouteHandler struct {
	sticker        *service.StickerService
	log            *log.Logger
	authMiddleware *authMw.JWTMiddleware
	validate       *utills.StructValidator
}

func NewStickerRouteHandler(sticker *service.StickerService, authMiddleware *authMw.JWTMiddleware, validate *utills.StructValidator) *StickerRouteHandler {
	return &StickerRouteHandler{
		sticker:        sticker,
		log:            log.New(os.Stdout, "StickerRouteHandler", log.LstdFlags|log.Lshortfile),
		authMiddleware: authMiddleware,
		validate:       validate,
	}
}

func (h *StickerRouteHandler) Mount(rg *gin.RouterGroup) {

	s0 := rg.Group("/check")
	s0.GET("/:id", h.authMiddleware.AuthRequired("user", "view"), g.InjectGin(h.checkSticker))

	s1 := rg.Group("/sets")
	s1.GET("/", h.authMiddleware.AuthRequired("user", "view"), g.InjectGin(h.listStickerSet))
	s1.POST("/find", g.InjectGin(h.findStickerSet))
	s1.POST("/create", h.authMiddleware.AuthRequired("admit", "add"), g.InjectGin(h.createStickerSet))

	s2 := rg.Group("/set")
	s2.GET("/:id", h.authMiddleware.AuthRequired("user", "view"), g.InjectGin(h.getStickerSet))
	s2.DELETE("/:id", g.InjectGin(h.removeStickerSet))
	s2.POST("/:id/add-sticker", h.authMiddleware.AuthRequired("user", "add"), g.InjectGin(h.addStickerToSet))
	s2.POST("/:id/remove-sticker", h.authMiddleware.AuthRequired("user", "edit"), g.InjectGin(h.removeStickerFromSet))
	s2.POST("/:id/edit", g.InjectGin(h.editStickerSet))

	s3 := rg.Group("/img")
	s3.GET("/:id", h.authMiddleware.AuthRequired("user", "view"), g.InjectGin(h.getStickerImage))

	s4 := rg.Group("/byid")
	s4.POST("/:id/edit", g.InjectGin(h.editSticker))

}

func (h *StickerRouteHandler) checkSticker(c *gin.Context, req struct{}) error {
	id := c.Param("id")
	if !bson.IsObjectIdHex(id) {
		return g.NewError(400, "bad param id")
	}

	exists, err := h.sticker.StickerExists(bson.ObjectIdHex(id))
	if err != nil {
		return err
	}
	c.JSON(200, g.Response{
		Success: true,
		Message: "",
		Data:    exists,
	})
	return nil
}

func (h *StickerRouteHandler) listStickerSet(c *gin.Context, req struct{}) error {
	stickers, err := h.sticker.ListStickerSet()
	if err != nil {
		return err
	}

	c.JSON(200, stickers)
	return nil
}

func (h *StickerRouteHandler) createStickerSet(c *gin.Context, req struct {
	Body dto.CreateStickerSetDto
}) error {
	id, err := h.sticker.NewStickerSet(req.Body)
	if err != nil {
		return err
	}
	c.JSON(200, gin.H{"id": id})
	return nil
}

func (h *StickerRouteHandler) getStickerSet(c *gin.Context, req struct{}) error {
	id := c.Param("id")
	if !bson.IsObjectIdHex(id) {
		return g.NewError(400, "bad param id")
	}

	stickers, err := h.sticker.GetStickerInSet(bson.ObjectIdHex(id))
	if err != nil {
		return err
	}

	c.JSON(200, stickers)
	return nil
}

var mapFormat = map[string]imaging.Format{
	"jpg":  imaging.JPEG,
	"jpeg": imaging.JPEG,
	"png":  imaging.PNG,
	"bmp":  imaging.BMP,
	"gif":  imaging.GIF,
	"tif":  imaging.TIFF,
	"tiff": imaging.TIFF,
}

func resizeImage(imgData []byte) ([]byte, error) {

	r := bytes.NewReader(imgData)
	// determine type
	_, format, err := image.DecodeConfig(r)
	if err != nil {
		log.Printf("image: error determining image type: %s, is it corrupt?", err)
		return nil, fmt.Errorf("image: error determining image type: %s, is it corrupt?", err)
	}
	//
	r.Seek(0, io.SeekStart) // reset seek
	src, err := imaging.Decode(r)
	if err != nil {
		log.Printf("imaging: error decoing image: %s, is it corrupt?", err)
		return nil, fmt.Errorf("imaging: error decoing image: %s, is it corrupt?", err)
	}

	size := src.Bounds().Size()
	width := size.X
	height := size.Y
	var img *image.NRGBA

	if width > 256 || height > 256 {
		if width > height {
			img = imaging.Resize(src, 256, 0, imaging.Lanczos)
		} else {
			img = imaging.Resize(src, 0, 256, imaging.Lanczos)
		}
	} else {
		img = imaging.Clone(src)
	}

	buf := new(bytes.Buffer)
	err = imaging.Encode(buf, img, mapFormat[format])
	if err != nil {
		log.Printf("error encoding to %s:%s\n", format, err)
		return nil, fmt.Errorf("error encoding to %s:%s\n", format, err)
	}

	resizedImage := buf.Bytes()
	return resizedImage, nil
}

func (h *StickerRouteHandler) addStickerToSet(c *gin.Context, req struct{}) error {
	setID := c.Param("id")
	if !bson.IsObjectIdHex(setID) {
		return g.NewError(400, "bad param id")
	}
	// TODO: manual bind CreateStickerDto

	name := c.PostForm("name")
	if name == "" {
		return g.NewError(400, "please specify name")
	}
	file, err := c.FormFile("image")
	if err != nil {
		return fmt.Errorf("error getting form file: %w", err)
	}

	f, err := file.Open()
	if err != nil {
		return fmt.Errorf("error opening file: %w", err)
	}
	defer f.Close()
	imageData, err := ioutil.ReadAll(f)
	if err != nil {
		return fmt.Errorf("error reading file: %w", err)
	}

	resized, err := resizeImage(imageData)
	if err != nil {
		return g.NewError(400, fmt.Sprintf("error resizing image: %s", err))
	}

	newId, err := h.sticker.AddStickerToSet(bson.ObjectIdHex(setID), dto.CreateStickerDto{Name: name}, resized)
	if err != nil {
		return err
	}

	c.JSON(200, gin.H{"id": newId})
	return nil
}

func (h *StickerRouteHandler) getStickerImage(c *gin.Context, req struct{}) error {
	stickerID := c.Param("id")
	if !bson.IsObjectIdHex(stickerID) {
		return g.NewError(400, "bad param id")
	}
	// TODO: manual bind CreateStickerDto

	image, err := h.sticker.GetStickerImage(bson.ObjectIdHex(stickerID))
	if err != nil {
		return err
	}

	// TODO[ROAD]: just let application decide 4 head?
	c.Data(200, "image/png", image)
	return nil
}

func (h *StickerRouteHandler) removeStickerFromSet(c *gin.Context, req struct {
	Body dto.RemoveStickersDto
}) error {

	errors := make([]string, 0, len(req.Body.IDs))
	for _, id := range req.Body.IDs {
		err := h.sticker.RemoveSticker(id)
		if err != nil {
			errors = append(errors, fmt.Sprintf("error deleting %s: %s", id.Hex(), err.Error()))
		}
	}

	c.JSON(200, gin.H{
		"errors": errors,
	})
	return nil
}

func (h *StickerRouteHandler) editStickerSet(c *gin.Context, req struct {
	Body dto.EditStickerSetDto
}) error {
	id := c.Param("id")
	if !bson.IsObjectIdHex(id) {
		return g.NewError(400, "bad param id")
	}

	if err := h.sticker.EditStickerSetInfo(bson.ObjectIdHex(id), req.Body); err != nil {
		return err
	}

	c.JSON(200, g.OK("updated sticker set"))
	return nil
}

func (h *StickerRouteHandler) editSticker(c *gin.Context, req struct {
	Body dto.EditStickerDto
}) error {
	id := c.Param("id")
	if !bson.IsObjectIdHex(id) {
		return g.NewError(400, "bad param roomId")
	}

	if err := h.sticker.EditStickerInfo(bson.ObjectIdHex(id), req.Body); err != nil {
		return err
	}

	c.JSON(200, g.OK("updated sticker"))
	return nil
}

func (h *StickerRouteHandler) removeStickerSet(c *gin.Context, req struct{}) error {
	id := c.Param("id")
	if !bson.IsObjectIdHex(id) {
		return g.NewError(400, "bad param stickerSet")
	}

	if empty, err := h.sticker.StickerSetIsEmpty(bson.ObjectIdHex(id)); err != nil {
		return err
	} else if !empty {
		return g.NewError(400, fmt.Sprintf("specified sticker set %s is not empty", id))
	}

	if err := h.sticker.RemoveStickerSet(bson.ObjectIdHex(id)); err != nil {
		return err
	}

	c.JSON(200, g.OK("removed sticker set"))
	return nil
}

func (h *StickerRouteHandler) findStickerSet(c *gin.Context, req struct {
	Body dto.FindStickerSetDto
}) error {
	stickers, err := h.sticker.FindStickerSet(req.Body)
	if err != nil {
		return err
	}
	c.JSON(200, stickers)
	return nil
}
