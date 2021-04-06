package plugin

import (
	"backendSenior/domain/model"
	"errors"
	"proxySenior/data/external"
)

// OnMessagePlugin is plugin for backup message
type OnMessagePortPlugin struct {
	enabled           bool
	enabledEncryption bool
	plugin            *external.GRPCOnPortMessagePlugin
}

// NewOnMessagePlugin create plugin from path of plugin server
func NewOnMessagePortPlugin(enabled bool, enabledEncrypt bool, serverAdd string) *OnMessagePortPlugin {

	if !enabled {
		return &OnMessagePortPlugin{
			enabled: false,
		}
	}

	return &OnMessagePortPlugin{
		enabled:           true,
		enabledEncryption: enabledEncrypt,
		plugin:            external.NewGRPCOnPortMessagePlugin(serverAdd),
	}
}

// IsEnabled return whether plugin is enabled
func (p *OnMessagePortPlugin) IsEnabled() bool {
	return p.enabled
}

// IsEnabled return whether plugin is enabled
func (p *OnMessagePortPlugin) IsEnabledEncryption() bool {
	return p.enabledEncryption
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
	//fmt.Printf("on message in: %v\nreturned: %v\n", message, err)
	return err
}

// OnMessageIn should be called when message in
func (p *OnMessagePortPlugin) CustomEncryptionPlugin(message model.Message) (model.Message, error) {
	if !p.enabled {
		return model.Message{}, errors.New("Plugin not enabled")
	}
	if !p.enabledEncryption {
		return model.Message{}, errors.New("Custom Encryption Plugin not enabled")
	}
	EncMessage, err := p.plugin.CustomEncryption(message)
	//fmt.Printf("on message in: %v\nreturned: %v\n", message, err)
	return EncMessage, err
}

// OnMessageIn should be called when message in
func (p *OnMessagePortPlugin) CustomDecryptionPlugin(message model.Message) (model.Message, error) {
	if !p.enabled {
		return model.Message{}, errors.New("Plugin not enabled")
	}

	if !p.enabledEncryption {
		return model.Message{}, errors.New("Custom Encryption Plugin not enabled")
	}

	DecMessage, err := p.plugin.CustomDecryption(message)
	//fmt.Printf("on message in: %v\nreturned: %v\n", message, err)
	return DecMessage, err
}

func (p *OnMessagePortPlugin) CloseConnection() {
	p.plugin.CloseConnection()
	return
}
