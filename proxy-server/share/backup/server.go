package backup

import (
	"context"
	"proxySenior/share/proto"
)

// BackupServer handle receiving request the forward it to `Impl` inside
// It's created when the server process run
type BackupServer struct {
	Impl BackupService
}

var _ proto.BackupServer = (*BackupServer)(nil)

// OnMessageIn delegate logic to `Impl`
func (s *BackupServer) OnMessageIn(ctx context.Context, req *proto.Chat) (*proto.Empty, error) {
	m := RawMessage{
		MessageID: req.MessageId,
		Data:      req.Data,
	}
	err := s.Impl.OnMessageIn(m)
	return &proto.Empty{}, err
}

// IsReady delegate logic to `Impl`
func (s *BackupServer) IsReady(ctx context.Context, req *proto.Empty) (*proto.Status, error) {
	ok, err := s.Impl.IsReady()
	return &proto.Status{
		Ok: ok,
	}, err
}
