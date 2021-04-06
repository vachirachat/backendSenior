package service

// import (
// 	"proxySenior/domain/service"
// )

// // WSRequestService provide HTTP request-response like communication
// // over the WebSocket
// type WSRequestService struct {
// 	// TODO: consider using pattern like this ?
// 	// msgIn chan([]byte)
// 	// msgOut chan([]byte)
// 	// then use channel to connect between service
// 	//
// 	up  *service.ChatUpstreamService
// 	req map[string](chan WSReqResRaw) // pending requests
// }

// // WSReqRes is to be embedded into the payload
// type WSReqResRaw struct {
// 	SenderID   string // proxyID of sender
// 	ReceiverID string // proxyID of receiver
// 	MessageID  string // uniqueID to determine transaction
// 	Type       string // type of message
// 	StatusCode int    // status code like http
// 	Payload    interface{}
// }

// // type WSPromise struct {
// // 	done   chan struct{} // will not block when request is done
// // 	data   interface{}
// // 	status int
// // }

// // New create new upstream service
// func New(up *service.ChatUpstreamService) *WSRequestService {
// 	return &WSRequestService{
// 		up: up,
// 	}
// }

// func (s *WSRequestService) Start() {
// 	pipe := make(chan []byte, 100)
// 	s.up.RegisterHandler(pipe)
// 	defer s.up.UnRegisterHandler(pipe)

// 	for {
// 		incData := <-pipe
// 		var incMessage chatsocket

// 	}
// }

// // Request make request, will return response
// func (s *WSRequestService) Request(req *WSReqResRaw) WSReqResRaw {
// 	//
// }
