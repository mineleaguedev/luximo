package models

type Plugin struct {
	Name        string   `json:"name"`
	Versions    []string `json:"versions"`
	LastVersion string   `json:"lastVersion"`
}

type PluginsResponse struct {
	Success bool     `json:"success"`
	Plugins []Plugin `json:"plugins"`
}
