package route

import (
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
	sticker  *service.StickerService
	log      *log.Logger
	validate *utills.StructValidator
}

func NewStickerRouteHandler(sticker *service.StickerService, validate *utills.StructValidator) *StickerRouteHandler {
	return &StickerRouteHandler{
		sticker:  sticker,
		log:      log.New(os.Stdout, "StickerRouteHandler", log.LstdFlags|log.Lshortfile),
		validate: validate,
	}
}

func (h *StickerRouteHandler) Mount(rg *gin.RouterGroup) {

	s0 := rg.Group("/check")
	s0.GET("/:id", g.InjectGin(h.checkSticker))

	s1 := rg.Group("/sets")
	s1.GET("/", g.InjectGin(h.listStickerSet))
	s1.POST("/create", g.InjectGin(h.createStickerSet))

	s2 := rg.Group("/set")
	s2.GET("/:id", g.InjectGin(h.getStickerSet))
	s2.POST("/:id/add-sticker", g.InjectGin(h.addStickerToSet))
	s2.POST("/:id/remove-sticker", g.InjectGin(h.removeStickerFromSet))

	s3 := rg.Group("/img")
	s3.GET("/:id", g.InjectGin(h.getStickerImage))

}

func (h *StickerRouteHandler) checkSticker(c *gin.Context, req struct {
}) error {
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

func (h *StickerRouteHandler) addStickerToSet(c *gin.Context, req struct {
}) error {
	setID := c.Param("id")
	if !bson.IsObjectIdHex(setID) {
		return g.NewError(400, "bad param id")
	}
	// TODO: manual bind CreateStickerDto

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

	newId, err := h.sticker.AddStickerToSet(bson.ObjectIdHex(setID), dto.CreateStickerDto{}, resized)
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

	panic("TODO")
	return nil
}
