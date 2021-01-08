package config

import (
	"proxySenior/share/backup"

	"github.com/hashicorp/go-plugin"
)

var HandshakeConfig = plugin.HandshakeConfig{
	ProtocolVersion:  1,
	MagicCookieKey:   "BACKUP",
	MagicCookieValue: "FOOBAR",
}

var PluginMaps = map[string]plugin.Plugin{
	"backup": &backup.BackupGRPCPlugin{},
}
