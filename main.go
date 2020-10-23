package main

import (
	"backendSenior/route"
	"log"

	"github.com/gin-gonic/gin"

	"github.com/globalsign/mgo"
)

const (
	mogoDBEnPint  = "mongodb://localhost:27017"
	portWebServie = ":3000"
)

func main() {
	connectionDB, err := mgo.Dial(mogoDBEnPint)
	if err != nil {
		log.Panic("Can no connect Database", err.Error())
	}
	router := gin.Default()
	route.NewRouteProduct(router, connectionDB)
	router.Run(portWebServie)
}
