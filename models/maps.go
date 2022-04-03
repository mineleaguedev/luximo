package models

type MiniGames struct {
	Name    string   `json:"name"`
	Formats []Format `json:"formats"`
}

type Format struct {
	Format string `json:"format"`
	Maps   []Map  `json:"maps"`
}

type Map struct {
	Name        string   `json:"name"`
	Versions    []string `json:"versions"`
	LastVersion string   `json:"lastVersion"`
	HasWorld    bool     `json:"hasWorld"`
	HasConfig   bool     `json:"hasConfig"`
}

type MapsResponse struct {
	Success   bool        `json:"success"`
	MiniGames []MiniGames `json:"minigames"`
}
