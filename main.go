package main

import (
	"github.com/mineleaguedev/luximo/handlers"
	"github.com/mineleaguedev/luximo/services"
	"github.com/spf13/viper"
	"github.com/valyala/fasthttp"
	"log"
)

func main() {
	viper.AddConfigPath(".")
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	if err := viper.ReadInConfig(); err != nil {
		log.Fatalf("Error reading config: %s", err.Error())
	}

	service := services.NewService()
	handler := handlers.NewHandler(service)

	r := handler.InitRoutes()
	log.Fatal(fasthttp.ListenAndServe(":8080", r.Handler))
}
