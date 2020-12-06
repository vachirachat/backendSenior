package api

import (
	"backendSenior/data/repository"
	"backendSenior/domain/model"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

type MessageAPI struct {
	MessageRepository repository.MessageRepository
}

func (api MessageAPI) MessageListHandler(context *gin.Context) {
	var messagesInfo model.MessageInfo
	messages, err := api.MessageRepository.GetAllMessage()
	if err != nil {
		log.Println("error MessageListHandler", err.Error())
		context.JSON(http.StatusInternalServerError, gin.H{"status": err.Error()})
		return
	}
	messagesInfo.Messages = messages
	context.JSON(http.StatusOK, messagesInfo)
}

func (api MessageAPI) GetMessageByIDHandler(context *gin.Context) {
	messageID := context.Param("message_id")
	message, err := api.MessageRepository.GetMessageByID(messageID)
	if err != nil {
		log.Println("error GetMessageByIDHandler", err.Error())
		context.JSON(http.StatusInternalServerError, gin.H{"status": err.Error()})
		return
	}
	context.JSON(http.StatusOK, message)
}

func (api MessageAPI) AddMessageHandeler(context *gin.Context) {
	var message model.Message
	err := context.ShouldBindJSON(&message)
	if err != nil {
		log.Println("error AddMessageHandeler", err.Error())
		context.JSON(http.StatusInternalServerError, gin.H{"status": err.Error()})
		return
	}
	err = api.MessageRepository.AddMessage(message)
	if err != nil {
		log.Println("error AddMessageHandeler", err.Error())
		context.JSON(http.StatusInternalServerError, gin.H{"status": err.Error()})
		return
	}
	context.JSON(http.StatusCreated, gin.H{"status": "success"})
}

func (api MessageAPI) DeleteMessageByIDHandler(context *gin.Context) {
	messageID := context.Param("message_id")
	err := api.MessageRepository.DeleteMessageByID(messageID)
	if err != nil {
		log.Println("error DeleteMessageHandler", err.Error())
		context.JSON(http.StatusInternalServerError, gin.H{"status": err.Error()})
	}
	context.JSON(http.StatusNoContent, gin.H{"status": "success"})
}
