package handlers

import (
	"encoding/json"
	"github.com/mineleaguedev/luximo/models"
	"github.com/valyala/fasthttp"
)

func (h *Handler) PluginsUpdateHandler(ctx *fasthttp.RequestCtx) {
	if err := h.services.UpdatePlugins(); err != nil {
		response, err := json.Marshal(&models.Error{
			Success: false,
			Message: err.Error(),
		})
		if err != nil {
			ctx.Error(err.Error(), 500)
			return
		}

		ctx.Error(string(response), 500)
		return
	}

	response, err := json.Marshal(&models.Response{
		Success: true,
	})
	if err != nil {
		ctx.Error(err.Error(), 500)
		return
	}

	_, err = ctx.WriteString(string(response))
	if err != nil {
		ctx.Error(err.Error(), 500)
		return
	}
}
