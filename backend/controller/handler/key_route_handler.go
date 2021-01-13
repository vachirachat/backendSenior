package route

import (
	"backendSenior/domain/service"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"

	"github.com/gin-gonic/gin"
	"github.com/globalsign/mgo/bson"
)

// KeyRoute is route for key sharing of controller
type KeyRoute struct {
	// keyAPI
	room *service.RoomService
}

func NewKeyRoute(room *service.RoomService) *KeyRoute {
	return &KeyRoute{
		room: room,
	}
}

func (r *KeyRoute) Mount(rg *gin.RouterGroup) {
	rg.POST("/room-key/:id", r.getRoomKeyFromProxy)
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
	fmt.Println("[get key] incoming request for", roomID)

	proxy, err := r.room.GetRoomMasterProxy(roomID)
	if err != nil {
		fmt.Println("keyRoute/getRoomKeyForProxy: can't get master proxy", err)
		c.JSON(400, gin.H{"status": "couldn't determine proxy to get key"})
		return
	}

	u := url.URL{
		Scheme: "http",
		Host:   proxy.IP + ":" + fmt.Sprint(proxy.Port),
		Path:   "/api/v1/key/" + roomID + "/get-key",
	}

	// cType := c.Request.Header.Get("content-type")
	// if cType == "" {
	// 	cTypeArr := c.Request.Header["content-type"]
	// 	if len(cTypeArr) > 0 {
	// 		cType = cTypeArr[0]
	// 	}
	// }

	// make request

	res, err := http.Post(u.String(), "application/json", c.Request.Body)
	if err != nil {
		fmt.Println("keyRoute/getRoomKeyForProxy: error making request to proxy", err)
		c.JSON(500, gin.H{"status": "error making request to proxy"})
		return
	}

	if res.StatusCode >= 400 {
		fmt.Println("keyRoute/getRoomKeyForProxy: proxy retured non OK status " + fmt.Sprint(res.StatusCode))
		c.JSON(500, gin.H{"status": "proxy retured non OK status " + fmt.Sprint(res.StatusCode)})
		return
	}

	body, err := ioutil.ReadAll(res.Body)
	defer res.Body.Close()
	fmt.Printf("[get key] proxy responded %s\n", body)

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
