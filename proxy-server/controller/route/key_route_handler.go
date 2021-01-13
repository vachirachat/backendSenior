package route

import (
	"backendSenior/domain/model/chatsocket/key_exchange"
	"fmt"
	"proxySenior/domain/service"

	"crypto/rsa"

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

	var keyReq key_exchange.KeyExchangeRequest
	c.ShouldBindJSON(&keyReq)

	var pk *rsa.PublicKey
	var shouldSendPK bool

	// it sends public key
	if len(keyReq.PublicKey) > 0 {
		shouldSendPK = true
		pk = h.k.BytesToPublicKey(keyReq.PublicKey)
		if pk == nil {
			c.JSON(400, gin.H{
				"status":  "error",
				"message": "the sent key isn't public key",
			})
		}
		h.k.SetProxyPublicKey(keyReq.ProxyID, keyReq.PublicKey)
	} else {
		localPk, ok := h.k.GetProxyPublicKey(keyReq.ProxyID)
		if !ok {
			c.JSON(500, gin.H{
				"status":  "error",
				"message": "no key for room, please send request with key",
			})
			return
		} else {
			pk = localPk
		}
	}

	keys, err := h.k.GetKeyLocal(roomID)
	if err != nil {
		fmt.Println("ERR keyRoute/getAll:", err)
		c.JSON(500, gin.H{"status": "error"})
		return
	}

	for i := range keys {
		// TODO: 2 layer encrypt
		// enc = encrypt with our private key
		enc := h.k.EncryptWithPublicKey(keys[i].Key, pk)
		if enc == nil {
			c.JSON(500, gin.H{
				"status":  "error",
				"message": "can't encrypt",
			})
			return
		}
		keys[i].Key = enc
	}

	var pkBytes []byte
	if shouldSendPK {
		pkBytes = h.k.PublicKeyToBytes(pk)
	}

	c.JSON(200, key_exchange.KeyExchangeResponse{
		PublicKey:    pkBytes,
		ProxyID:      keyReq.ProxyID,
		RoomID:       keyReq.RoomID,
		Keys:         keys,
		ErrorMessage: "[Note] using old key",
	})
}
