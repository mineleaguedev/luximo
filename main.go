package main

import (
	"github.com/mineleaguedev/luximo/handlers"
	"github.com/mineleaguedev/luximo/models"
	"github.com/mineleaguedev/luximo/services"
	"github.com/spf13/viper"
	"github.com/valyala/fasthttp"
	"log"
	"os"
)

func main() {
	viper.AddConfigPath(".")
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	if err := viper.ReadInConfig(); err != nil {
		log.Fatalf("Error reading config: %s", err.Error())
	}

	paths := models.Paths{
		Path: viper.GetString("path"),
	}
	paths.VelocityPath = paths.Path + viper.GetString("velocity_path")
	paths.PaperPath = paths.Path + viper.GetString("paper_path")
	paths.PluginsPath = paths.Path + viper.GetString("plugins_path")
	paths.MapsPath = paths.Path + viper.GetString("maps_path")
	paths.ServersPath = paths.Path + viper.GetString("servers_path")
	paths.LobbyServersPath = paths.ServersPath + viper.GetString("lobby_servers_path")
	paths.MiniServersPath = paths.ServersPath + viper.GetString("mini_servers_path")
	paths.MegaServersPath = paths.ServersPath + viper.GetString("mega_servers_path")
	if err := os.MkdirAll(paths.VelocityPath, 0755); err != nil {
		log.Fatalf("Error creating velocity folder: %s", err.Error())
	}
	if err := os.MkdirAll(paths.PaperPath, 0755); err != nil {
		log.Fatalf("Error creating paper folder: %s", err.Error())
	}
	if err := os.MkdirAll(paths.PluginsPath, 0755); err != nil {
		log.Fatalf("Error creating plugins folder: %s", err.Error())
	}
	if err := os.MkdirAll(paths.MapsPath, 0755); err != nil {
		log.Fatalf("Error creating maps folder: %s", err.Error())
	}
	if err := os.MkdirAll(paths.LobbyServersPath, 0755); err != nil {
		log.Fatalf("Error creating lobby servers folder: %s", err.Error())
	}
	if err := os.MkdirAll(paths.MiniServersPath, 0755); err != nil {
		log.Fatalf("Error creating mini servers folder: %s", err.Error())
	}
	if err := os.MkdirAll(paths.MegaServersPath, 0755); err != nil {
		log.Fatalf("Error creating mega servers folder: %s", err.Error())
	}

	service := services.NewService(paths)
	handler := handlers.NewHandler(service)

	r := handler.InitRoutes()
	log.Fatal(fasthttp.ListenAndServe(":8080", r.Handler))
}
