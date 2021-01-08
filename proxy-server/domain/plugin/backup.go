package plugin

import (
	"backendSenior/domain/model"
	"errors"
	"fmt"
	"os/exec"
	"proxySenior/data/external"
)

// OnMessagePlugin is plugin for backup message
type OnMessagePlugin struct {
	enabled bool
	plugin  *external.GRPCOnMessagePlugin
}

// NewOnMessagePlugin create plugin from path of plugin server
func NewOnMessagePlugin(enabled bool, path string) *OnMessagePlugin {
	if !enabled {
		return &OnMessagePlugin{
			enabled: false,
		}
	}
	return &OnMessagePlugin{
		enabled: true,
		plugin: external.NewGRPCOnMessagePlugin(
			exec.Command("sh", "-c", path),
		),
	}
}

// IsEnabled return whether plugin is enabled
func (p *OnMessagePlugin) IsEnabled() bool {
	return p.enabled
}

// Wait blocks until plugin is ready,
func (p *OnMessagePlugin) Wait() error {
	if !p.enabled {
		return errors.New("Plugin not enabled")
	}
	return p.plugin.Wait()
}

// OnMessageIn should be called when message in
func (p *OnMessagePlugin) OnMessageIn(message model.Message) error {
	if !p.enabled {
		return errors.New("Plugin not enabled")
	}
	err := p.plugin.OnMessageIn(message)
	fmt.Printf("on message in: %v\nreturned: %v\n", message, err)
	return err
}
