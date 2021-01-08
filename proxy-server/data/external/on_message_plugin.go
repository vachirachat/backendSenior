package external

import (
	"backendSenior/domain/model"
	"fmt"
	"log"
	"os/exec"
	"proxySenior/share/backup"
	"proxySenior/share/config"
	"time"

	"github.com/hashicorp/go-plugin"
)

// GRPCOnMessagePlugin is struct for plugins over GRPC
type GRPCOnMessagePlugin struct {
	client        *plugin.Client
	backupService backup.BackupService
}

// NewGRPCOnMessagePlugin create new GRPC plugin that is intended to be called when recv message
// parameter cmd is command to be exec to run the plugin
// example is exec.Cmd("sh", "-c", "./plugin")
func NewGRPCOnMessagePlugin(cmd *exec.Cmd) *GRPCOnMessagePlugin {

	client := plugin.NewClient(&plugin.ClientConfig{
		HandshakeConfig: config.HandshakeConfig,
		Plugins:         config.PluginMaps,
		AllowedProtocols: []plugin.Protocol{
			plugin.ProtocolGRPC,
		},
		Cmd: cmd,
	})

	grpcClient, err := client.Client()
	if err != nil {
		log.Fatalln("err", err)
	}

	raw, err := grpcClient.Dispense("backup")
	if err != nil {
		log.Fatalln("err", err)
	}

	svc, ok := raw.(backup.BackupService)
	if !ok {
		log.Fatal("Assertion failed")
	}

	return &GRPCOnMessagePlugin{
		client:        client,
		backupService: svc,
	}
}

// Wait blocks until underlying GRPC server is ready
func (p *GRPCOnMessagePlugin) Wait() error {
	fmt.Println("waiting for GRPC server...")
	for {
		ok, err := p.backupService.IsReady()
		if err != nil {
			return err
		}
		if ok {
			return nil
		}
		time.Sleep(1 * time.Second)
	}
}

// GetService return instance of backup service to be called
func (p *GRPCOnMessagePlugin) GetService() backup.BackupService {
	return p.backupService
}

// OnMessageIn convert message from model.Message then send over GRPC
func (p *GRPCOnMessagePlugin) OnMessageIn(message model.Message) error {
	fmt.Println("[plugin] message is", message)
	err := p.backupService.OnMessageIn(backup.RawMessage{
		MessageID: message.MessageID.Hex(),
		TimeStamp: message.TimeStamp.Unix(),
		RoomID:    message.RoomID.Hex(),
		UserID:    message.UserID.Hex(),
		ClientUID: message.ClientUID,
		Data:      message.Data,
		Type:      message.Type,
	})
	return err
}
