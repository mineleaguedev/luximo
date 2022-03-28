package services

import (
	"github.com/mineleaguedev/luximo/models"
)

type Velocity interface {
}

type Paper interface {
}

type Plugin interface {
	UpdatePlugins() error
	GetPluginsInfo() ([]models.Plugin, error)
	DownloadPlugin(pluginName, version string) (*[]byte, error)
	UpdatePlugin(pluginName, version string, pluginFileBytes []byte) error
	DeletePlugin(pluginName string) error
}

type Map interface {
}

type ProxyServer interface {
}

type LobbyServer interface {
}

type MiniServer interface {
}

type MegaServer interface {
}

type Service struct {
	Velocity
	Paper
	Plugin
	Map
	ProxyServer
	LobbyServer
	MiniServer
	MegaServer
}

func NewService(paths models.Paths) *Service {
	return &Service{
		Plugin: NewPluginService(paths),
	}
}
