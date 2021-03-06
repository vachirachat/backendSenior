package route

import (
	"backendSenior/domain/model/chatsocket/key_exchange"
	"fmt"
	"proxySenior/domain/encryption"
	"proxySenior/domain/service/key_service"

	"crypto/rsa"

	"github.com/gin-gonic/gin"
	"github.com/globalsign/mgo/bson"
)

// KeyRoute route for
// - getting and generating key
type KeyRoute struct {
	k *key_service.KeyService
}

// NewKeyRouteHandler create new handler
func NewKeyRouteHandler(keyService *key_service.KeyService) *KeyRoute {
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
	roomID := c.Param("id")
	if roomID == "" || !bson.IsObjectIdHex(roomID) {
		c.JSON(400, gin.H{"status": "bad room ID"})
		return
	}

	var keyReq key_exchange.KeyExchangeRequest
	var err error
	err = c.ShouldBindJSON(&keyReq)
	if err != nil {
		fmt.Println("error", err)
		c.JSON(400, gin.H{"status": "bad request"})
		return
	}

	var pk *rsa.PublicKey
	var shouldSendPK bool

	// it sends public key
	if len(keyReq.PublicKey) > 0 {
		shouldSendPK = true
		pk, err = encryption.BytesToPublicKey(keyReq.PublicKey)
		if pk == nil {
			c.JSON(400, gin.H{
				"status":  "error",
				"message": "public key error:" + err.Error(),
			})
			return
		}
		// remember key
		//fmt.Printf("[key-exchange] request: remembering proxy %s with key\n%s\n", keyReq.ProxyID, keyReq.PublicKey)
		fmt.Printf("[key-exchange] request: remembering public sent from proxy %s\n", keyReq.ProxyID)
		h.k.SetProxyPublicKey(keyReq.ProxyID, pk)
	} else {
		localPk, ok := h.k.GetProxyPublicKey(keyReq.ProxyID)
		if !ok {
			c.JSON(500, key_exchange.KeyExchangeResponse{
				PublicKey:    nil,
				ProxyID:      keyReq.ProxyID,
				RoomID:       keyReq.RoomID,
				Keys:         nil,
				ErrorMessage: "ERR_NO_KEY",
			})
			return
		} else {
			pk = localPk
		}
		fmt.Printf("[key-exchange] encrypt with public key of %s\n%s\n", keyReq.ProxyID, encryption.PublicKeyToBytes(localPk))
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
		enc, err := encryption.EncryptWithPublicKey(keys[i].Key, pk)
		if enc == nil {
			// we should return with res
			c.JSON(500, gin.H{
				"status":  "error",
				"message": "error encryption: " + err.Error(),
			})
			return
		}
		keys[i].Key = enc
	}

	var pkBytes []byte
	if shouldSendPK {
		// TODO should send my pk ?
		pkBytes = h.k.MyKeyBytes()
	}

	message := ""
	if !shouldSendPK {
		message = "USING OLD KEY"
	}

	c.JSON(200, key_exchange.KeyExchangeResponse{
		PublicKey:    pkBytes,
		ProxyID:      keyReq.ProxyID,
		RoomID:       keyReq.RoomID,
		Keys:         keys,
		ErrorMessage: message,
	})
}
