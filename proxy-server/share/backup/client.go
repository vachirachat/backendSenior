package backup

// import (
// 	"context"
// 	"proxySenior/share/proto"
// )

// // backup client wrap backup service
// type BackupClient struct {
// 	c proto.BackupClient
// }

// var _ BackupService = (*BackupClient)(nil)

// func (c *BackupClient) OnMessageIn(m RawMessage) error {
// 	_, err := c.c.OnMessageIn(context.TODO(), &proto.Chat{
// 		MessageId: m.MessageID,
// 		RoomId:    m.RoomID,
// 		UserId:    m.UserID,
// 		Timestamp: m.TimeStamp,
// 		Type:      m.Type,
// 		ClientUid: m.ClientUID,
// 		Data:      m.Data,
// 	})
// 	return err
// }

// func (c *BackupClient) IsReady() (bool, error) {
// 	res, err := c.c.IsReady(context.TODO(), &proto.Empty{})
// 	return res.Ok, err
// }
