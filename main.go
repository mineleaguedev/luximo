package main

import (
	"github.com/mineleaguedev/luximo/handlers"
	"github.com/mineleaguedev/luximo/services"
	"github.com/valyala/fasthttp"
	"log"
)

func main() {
	service := services.NewService()
	handler := handlers.NewHandler(service)

	r := handler.InitRoutes()
	log.Fatal(fasthttp.ListenAndServe(":8080", r.Handler))
}
