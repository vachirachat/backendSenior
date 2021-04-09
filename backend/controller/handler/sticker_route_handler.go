package route

import (
	"backendSenior/domain/dto"
	"backendSenior/domain/service"
	g "common/utils/ginutils"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/globalsign/mgo/bson"
	"io/ioutil"
	"log"
	"os"
)

type StickerRouteHandler struct {
	sticker *service.StickerService
	log     *log.Logger
}

func NewStickerRouteHandler(sticker *service.StickerService) *StickerRouteHandler {
	return &StickerRouteHandler{
		sticker: sticker,
		log:     log.New(os.Stdout, "StickerRouteHandler", log.LstdFlags|log.Lshortfile),
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

}

func (h *StickerRouteHandler) checkSticker(c *gin.Context, req struct {
}) error {
	id := c.Param("id")
	if !bson.IsObjectIdHex(id) {
		return g.NewError(400, "bad room id")
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

func (h *StickerRouteHandler) addStickerToSet(c *gin.Context, req struct {
	Body dto.CreateStickerDto
}) error {
	setID := c.Param("id")
	if !bson.IsObjectIdHex(setID) {
		return g.NewError(400, "bad param id")
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
	bytes, err := ioutil.ReadAll(f)
	if err != nil {
		return fmt.Errorf("error reading file: %w", err)
	}

	newId, err := h.sticker.AddStickerToSet(bson.ObjectIdHex(setID), req.Body, bytes)
	if err != nil {
		return err
	}

	c.JSON(200, gin.H{"id": newId})
	return nil
}

func (h *StickerRouteHandler) removeStickerFromSet(c *gin.Context, req struct {
	Body dto.RemoveStickersDto
}) error {

	panic("TODO")
	return nil
}
