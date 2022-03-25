package services

type Velocity interface {
}

type Paper interface {
}

type Plugin interface {
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

func NewService() *Service {
	return &Service{}
}
