package plugin

import (
	"backendSenior/domain/model"
	"errors"
	"fmt"
	"proxySenior/data/external"
)

// OnMessagePlugin is plugin for backup message
type OnMessagePortPlugin struct {
	enabled bool
	plugin  *external.GRPCOnPortMessagePlugin
}

// NewOnMessagePlugin create plugin from path of plugin server
func NewOnMessagePortPlugin(enabled bool, serverAdd string) *OnMessagePortPlugin {

	if !enabled {
		return &OnMessagePortPlugin{
			enabled: false,
		}
	}
	return &OnMessagePortPlugin{
		enabled: true,
		plugin:  external.NewGRPCOnPortMessagePlugin(serverAdd),
	}
}

// IsEnabled return whether plugin is enabled
func (p *OnMessagePortPlugin) IsEnabled() bool {
	return p.enabled
}

// Wait blocks until plugin is ready,
func (p *OnMessagePortPlugin) Wait() error {
	if !p.enabled {
		return errors.New("Plugin not enabled")
	}
	return p.plugin.Wait()
}

// OnMessageIn should be called when message in
func (p *OnMessagePortPlugin) OnMessagePortPlugin(message model.Message) error {
	if !p.enabled {
		return errors.New("Plugin not enabled")
	}
	err := p.plugin.OnMessageIn(message)
	fmt.Printf("on message in: %v\nreturned: %v\n", message, err)
	return err
}

func (p *OnMessagePortPlugin) CloseConnection() {
	p.plugin.CloseConnection()
	return
}
