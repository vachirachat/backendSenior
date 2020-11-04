package api

import (
	"backendSenior/repository/pubsub"

	"github.com/gin-gonic/gin"
)

func handleSession(context *gin.Context) {
	hub := pubsub.H
	go hub.Run()
	pubsub.ServeWs(context)
}
