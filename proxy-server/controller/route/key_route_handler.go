package route

import (
	"fmt"
	"proxySenior/domain/service"

	"github.com/gin-gonic/gin"
	"github.com/globalsign/mgo/bson"
)

// KeyRoute route for
// - getting and generating key
type KeyRoute struct {
	k *service.KeyService
}

// NewKeyRouteHandler create new handler
func NewKeyRouteHandler(keyService *service.KeyService) *KeyRoute {
	return &KeyRoute{
		k: keyService,
	}
}

// Mount add routes to router group
func (h *KeyRoute) Mount(rg *gin.RouterGroup) {
	rg.POST("/:id/key", h.generate)
	rg.POST("/:id/get-key", h.getAll) // it's post since it require more data
}

func (h *KeyRoute) generate(c *gin.Context) {
	// TODO check if is local
	roomID := c.Param("id")
	if roomID == "" || !bson.IsObjectIdHex(roomID) {
		c.JSON(400, gin.H{"status": "bad room ID"})
		return
	}

	err := h.k.NewKeyForRoom(roomID)
	if err != nil {
		fmt.Println("ERR keyRoute/generate: ", err)
		c.JSON(500, gin.H{"status": "error"})
		return
	}

	c.JSON(200, gin.H{"status": "OK"})
}

func (h *KeyRoute) getAll(c *gin.Context) {
	// TODO: when implement publickey encryption, determine the requester proxy to use public key
	roomID := c.Param("id")
	if roomID == "" || !bson.IsObjectIdHex(roomID) {
		c.JSON(400, gin.H{"status": "bad room ID"})
		return
	}

	keys, err := h.k.GetKeyLocal(roomID)
	if err != nil {
		fmt.Println("ERR keyRoute/getAll:", err)
		c.JSON(500, gin.H{"status": "error"})
		return
	}

	c.JSON(200, keys)
}
