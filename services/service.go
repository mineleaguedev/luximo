package services

import (
	"github.com/mineleaguedev/luximo/models"
)

type Velocity interface {
	UpdateVelocity() error
	GetVelocityVersionsInfo() (*models.VelocityResponse, error)
	DownloadVelocity(version string) (*[]byte, error)
	UpdateVelocityVersion(version string, velocityFileBytes []byte) error
}

type Paper interface {
	UpdatePaper() error
	GetPaperVersionsInfo() (*models.PaperResponse, error)
	DownloadPaper(version string) (*[]byte, error)
	UpdatePaperVersion(version string, paperFileBytes []byte) error
}

type Plugin interface {
	UpdatePlugins() error
	GetPluginsInfo() ([]models.Plugin, error)
	DownloadPlugin(pluginName, version string) (*[]byte, error)
	UpdatePlugin(pluginName, version string, pluginFileBytes []byte) error
}

type Map interface {
	UpdateMaps() error
	GetMapsInfo() ([]models.MiniGames, error)
	DownloadMapWorld(minigame, format, minigameMap, version string) (*[]byte, error)
	DownloadMapConfig(minigame, format, minigameMap, version string) (*[]byte, error)
	UpdateMap(minigame, format, mapName, version string, mapWorldFileBytes, mapConfigFileBytes *[]byte) error
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
		Plugin:   NewPluginService(paths),
		Map:      NewMapService(paths),
		Velocity: NewVelocityService(paths),
		Paper:    NewPaperService(paths),
	}
}
