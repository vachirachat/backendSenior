package plugin

import (
	"backendSenior/domain/model"
	"errors"
	"fmt"
	"proxySenior/data/external"
	model_proxy "proxySenior/domain/model"
)

// OnMessagePlugin is plugin for backup message
type OnMessagePortPlugin struct {
	proxyConfig *model_proxy.ProxyConfig
	plugin      *external.GRPCOnPortMessagePlugin
}

// NewOnMessagePlugin create plugin from path of plugin server
func NewOnMessagePortPlugin(proxyConfig *model_proxy.ProxyConfig) *OnMessagePortPlugin {

	// if !proxyConfig.EnablePlugin {
	// 	return &OnMessagePortPlugin{
	// 		proxyConfig: proxyConfig,
	// 		plugin:      external.NewGRPCOnPortMessagePlugin(proxyConfig),
	// 	}
	// }

	return &OnMessagePortPlugin{
		proxyConfig: proxyConfig,
		plugin:      external.NewGRPCOnPortMessagePlugin(proxyConfig),
	}
}

// IsEnabled return whether plugin is enabled
func (p *OnMessagePortPlugin) IsEnabled() bool {
	return p.proxyConfig.EnablePlugin
}

// IsEnabled return whether plugin is enabled
func (p *OnMessagePortPlugin) IsEnabledEncryption() bool {
	return p.proxyConfig.EnablePluginEnc
}

// Wait blocks until plugin is ready,
func (p *OnMessagePortPlugin) Wait() error {
	if !p.proxyConfig.EnablePlugin {
		return errors.New("Plugin not enabled")
	}
	return p.plugin.Wait()
}

// OnMessageIn should be called when message in
func (p *OnMessagePortPlugin) OnMessagePortPlugin(message model.Message) error {
	if !p.proxyConfig.EnablePlugin {
		return errors.New("Plugin not enabled")
	}
	err := p.plugin.OnMessageIn(message)
	fmt.Printf("on message in: %v\nreturned: %v\n", message, err)
	return err
}

// OnMessageIn should be called when message in
func (p *OnMessagePortPlugin) CustomEncryptionPlugin(message model.Message) (model.Message, error) {
	if !p.proxyConfig.EnablePlugin {
		return model.Message{}, errors.New("Plugin not enabled")
	}
	if !p.proxyConfig.EnablePluginEnc {
		return model.Message{}, errors.New("Custom Encryption Plugin not enabled")
	}
	EncMessage, err := p.plugin.CustomEncryption(message)
	fmt.Printf("on message in: %v\nreturned: %v\n", message, err)
	return EncMessage, err
}

// OnMessageIn should be called when message in
func (p *OnMessagePortPlugin) CustomDecryptionPlugin(message model.Message) (model.Message, error) {
	if !p.proxyConfig.EnablePlugin {
		return model.Message{}, errors.New("Plugin not enabled")
	}

	if !p.proxyConfig.EnablePluginEnc {
		return model.Message{}, errors.New("Custom Encryption Plugin not enabled")
	}

	DecMessage, err := p.plugin.CustomDecryption(message)
	fmt.Printf("on message in: %v\nreturned: %v\n", message, err)
	return DecMessage, err
}

func (p *OnMessagePortPlugin) CloseConnection() {
	p.plugin.CloseConnection()
	return
}
