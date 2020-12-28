package chatsocket

import "backendSenior/domain/model/chatsocket"

type readCmd struct {
	userID string
	//  return value
	// optimize: return raw data first, outer thread will process later
	result chan []*chatsocket.Connection
	err    chan error
}

type addCmd struct {
	conn *chatsocket.Connection
	//  return value
	err chan error
}

type deleteCmd struct {
	connID string
	//  return value
	err chan error
}
