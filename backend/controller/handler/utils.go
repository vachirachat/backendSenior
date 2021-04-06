package route

import (
	"github.com/gin-gonic/gin"
	"github.com/globalsign/mgo/bson"
)

func mustGetObjectID(c *gin.Context, name string) (string, bool) {
	id := c.Param(name)
	if id == "" || !bson.IsObjectIdHex(id) {
		c.JSON(400, gin.H{
			"status":  "error",
			"message": "bad param " + name,
		})
		return "", false
	}
	return id, true
}
