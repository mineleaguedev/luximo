package handlers

import (
	"github.com/fasthttp/router"
	"github.com/mineleaguedev/luximo/services"
)

type Handler struct {
	services *services.Service
}

func NewHandler(services *services.Service) *Handler {
	return &Handler{
		services: services,
	}
}

func (h *Handler) InitRoutes() *router.Router {
	r := router.New()
	return r
}
