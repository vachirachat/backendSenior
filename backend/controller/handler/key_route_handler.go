package route

import (
	"backendSenior/domain/model/chatsocket"
	"backendSenior/domain/service"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"

	"github.com/gin-gonic/gin"
	"github.com/globalsign/mgo/bson"
)

// KeyRoute is route for key sharing of controller
type KeyRoute struct {
	// keyAPI
	proxy *service.ProxyService
	keyEx *service.KeyExchangeService
	chat  *service.ChatService // for broadcast message to room when key change
}

func NewKeyRoute(proxy *service.ProxyService, keyEx *service.KeyExchangeService, chat *service.ChatService) *KeyRoute {
	return &KeyRoute{
		proxy: proxy,
		keyEx: keyEx,
		chat:  chat,
	}
}

func (r *KeyRoute) Mount(rg *gin.RouterGroup) {
	rg.POST("/room-key/:id", r.getRoomKeyFromProxy)
	rg.POST("/room-key/:id/generate", r.generateRoomKey)
	rg.GET("/master-proxy/:id", r.getMasterProxy) // return *current* master proxy

	rg.GET("/priority/:roomId", r.getRoomPriority)
	rg.POST("/priority/:roomId/:proxyId", r.setRoomPriority)
	rg.POST("/catch-up/:roomId/:proxyId", r.catchUpKeyVersion) // proxy tell controller that its version of key is updated

	rg.GET("/public-key/:id")
	rg.POST("/public-key/:id")
}

// getRoomKeyFromProxy this just proxy-pass the request, it doesn't parse request in anyway
func (r *KeyRoute) getRoomKeyFromProxy(c *gin.Context) {
	roomID := c.Param("id")
	if !bson.IsObjectIdHex(roomID) {
		c.JSON(400, gin.H{"status": "bad room id"})
		return
	}

	pid, err := r.keyEx.GetMaster(roomID)
	if err != nil {
		fmt.Println("keyRoute/getRoomKeyForProxy: can't get master proxy", err)
		c.JSON(500, gin.H{"status": err.Error()})
		return
	}

	proxy, err := r.proxy.GetProxyByID(pid)
	if err != nil {
		fmt.Println("keyRoute/getRoomKeyForProxy: can't get master proxy", err)
		c.JSON(500, gin.H{"status": "couldn't determine proxy to get key"})
		return
	}

	u := url.URL{
		Scheme: "http",
		Host:   proxy.IP + ":" + fmt.Sprint(proxy.Port),
		Path:   "/api/v1/key/" + roomID + "/get-key",
	}
	// make request

	res, err := http.Post(u.String(), "application/json", c.Request.Body)
	if err != nil {
		fmt.Println("keyRoute/getRoomKeyForProxy: error making request to proxy", err)
		c.JSON(500, gin.H{"status": "error making request to proxy"})
		return
	}

	if res.StatusCode >= 400 {
		body, _ := ioutil.ReadAll(res.Body)
		fmt.Printf("keyRoute/getRoomKeyForProxy: proxy retured non OK status %d\nbody%s\n", res.StatusCode, body)
		c.Data(res.StatusCode, res.Header.Get("Content-Type"), body)
		return
	}

	body, err := ioutil.ReadAll(res.Body)
	defer res.Body.Close()
	// fmt.Printf("[get key] proxy responded %s\n", body)

	// TODO: do we need to verify ? or just pass the response ?
	var resBody interface{}
	err = json.Unmarshal(body, &resBody)
	if err != nil {
		fmt.Println("keyRoute/getRoomKeyForProxy: error decoding proxy response", err)
		c.JSON(500, gin.H{"status": "error decoding proxy response"})
		return
	}

	// dupe response

	// cType = c.Request.Header.Get("content-type")
	// if cType == "" {
	// 	cTypeArr := c.Request.Header["content-type"]
	// 	if len(cTypeArr) > 0 {
	// 		cType = cTypeArr[0]
	// 	}
	// }

	c.JSON(200, resBody)
}

// generateRoomKey tell proxy to generate key
func (r *KeyRoute) generateRoomKey(c *gin.Context) {
	roomID := c.Param("id")
	if !bson.IsObjectIdHex(roomID) {
		c.JSON(400, gin.H{"status": "bad room id"})
		return
	}
	fmt.Println("[get key] incoming request for", roomID)

	pid, err := r.keyEx.GetMaster(roomID)
	if err != nil {
		fmt.Println("keyRoute/generateRoomKey: can't get master proxy", err)
		c.JSON(500, gin.H{"status": err.Error()})
		return
	}

	proxy, err := r.proxy.GetProxyByID(pid)
	if err != nil {
		fmt.Println("keyRoute/generateRoomKey: can't get master proxy", err)
		c.JSON(500, gin.H{"status": "couldn't determine proxy to get key"})
		return
	}

	u := url.URL{
		Scheme: "http",
		Host:   proxy.IP + ":" + fmt.Sprint(proxy.Port),
		Path:   "/api/v1/key/" + roomID + "/key",
	}
	// make request

	res, err := http.Post(u.String(), "application/json", nil)
	if err != nil {
		fmt.Println("keyRoute/generateRoomKey: error making request to proxy", err)
		c.JSON(500, gin.H{"status": "error making request to proxy"})
		return
	}

	// Proxy responded, but non-ok, should forward messasge to requester
	if res.StatusCode >= 400 {
		body, _ := ioutil.ReadAll(res.Body)
		log.Printf("proxy returned non ok status: %d\nbody\n", res.StatusCode, body)
		c.Data(res.StatusCode, res.Header.Get("Content-Type"), body)
		return
	}

	err = r.keyEx.IncrementVersion(roomID, proxy.ProxyID.Hex())
	if err != nil {
		fmt.Println("keyRoute/generateRoomKey: increment version error ", err)
		c.JSON(500, gin.H{"status": "error"})
	}

	go r.chat.BroadcastMessageToRoom(roomID, chatsocket.InvalidateRoomKeyMessage(roomID))

	body, err := ioutil.ReadAll(res.Body)
	defer res.Body.Close()
	_ = body
	// fmt.Printf("[get key] proxy responded %s\n", body)

	c.JSON(200, gin.H{"status": "OK"})
}

// getMasterProxy return current master proxy of room
func (r *KeyRoute) getMasterProxy(c *gin.Context) {
	roomID := c.Param("id")
	if !bson.IsObjectIdHex(roomID) {
		c.JSON(400, gin.H{"status": "bad room id"})
		return
	}
	fmt.Println("[get key] incoming request for", roomID)

	pid, err := r.keyEx.GetMaster(roomID)
	if err != nil {
		fmt.Println("keyRoute/getMasterProxy: can't get master proxy", err)
		c.JSON(500, gin.H{"status": err.Error()})
		return
	}

	proxy, err := r.proxy.GetProxyByID(pid)
	if err != nil {
		fmt.Println("keyRoute/getMasterProxy: can't get master proxy", err)
		c.JSON(500, gin.H{"status": "couldn't determine proxy to get key"})
		return
	}

	c.JSON(200, proxy)
}

func (r *KeyRoute) setRoomPriority(c *gin.Context) {
	roomID := c.Param("roomId")
	proxyID := c.Param("proxyId")
	if !bson.IsObjectIdHex(roomID) || !bson.IsObjectIdHex(proxyID) {
		c.JSON(400, gin.H{"status": "bad room id or proxy id"})
		return
	}

	var body struct {
		Priority *int `json:"priority"`
	}
	err := c.ShouldBindJSON(&body)
	if err != nil || body.Priority == nil {
		c.JSON(400, gin.H{"status": "bad or invalid `priority` field"})
		return
	}

	err = r.keyEx.SetPriority(roomID, proxyID, *body.Priority)
	if err != nil {
		fmt.Println("key/setRoomPriority: err", err)
		c.JSON(500, gin.H{"status": "error"})
		return
	}

	c.JSON(200, gin.H{"status": "OK"})
}

func (r *KeyRoute) getRoomPriority(c *gin.Context) {
	roomID := c.Param("roomId")
	if !bson.IsObjectIdHex(roomID) {
		c.JSON(400, gin.H{"status": "bad room id"})
		return
	}

	priorities, err := r.keyEx.GetPriorities(roomID)

	if err != nil {
		fmt.Println("key/getRoomPriority: err", err)
		c.JSON(500, gin.H{"status": "error"})
		return
	}

	c.JSON(200, priorities)
}

func (r *KeyRoute) catchUpKeyVersion(c *gin.Context) {
	roomID := c.Param("roomId")
	proxyID := c.Param("proxyId")
	if !bson.IsObjectIdHex(roomID) || !bson.IsObjectIdHex(proxyID) {
		c.JSON(400, gin.H{"status": "bad room id or proxy id"})
		return
	}

	err := r.keyEx.CatchupKeyVersion(roomID, proxyID)
	if err != nil {
		fmt.Println("key/catchUpKeyVersion: error updating version", err)
		c.JSON(500, gin.H{"status": "error updating version, try again"})
		return
	}

	// when catchup, master could change, should invalidate old one
	go r.chat.BroadcastMessageToRoom(roomID, chatsocket.InvalidateRoomMasterMessage(roomID))

	c.JSON(200, gin.H{"stauts": "OK"})
}
