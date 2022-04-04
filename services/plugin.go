package services

import (
	"encoding/json"
	"errors"
	"github.com/hashicorp/go-version"
	"github.com/mineleaguedev/luximo/models"
	"io/ioutil"
	"net/http"
	"os"
	"sort"
	"strings"
)

type PluginService struct {
	paths models.Paths
}

func NewPluginService(paths models.Paths) *PluginService {
	return &PluginService{paths: paths}
}

func (s *PluginService) UpdatePlugins() error {
	var pluginsArr []models.Plugin

	pluginsInfo, err := s.GetPluginsInfo()
	if err != nil {
		return err
	}

	plugins, err := os.ReadDir(s.paths.PluginsPath)
	if err != nil {
		return err
	}

	if len(plugins) == 0 {
		if err := s.deleteOldAndWrongPlugins(pluginsArr, pluginsInfo); err != nil {
			return err
		}

		if err := s.addNewPlugins(pluginsInfo); err != nil {
			return err
		}

		return nil
	}

	for _, plugin := range plugins {
		if !strings.Contains(plugin.Name(), ".jar") {
			if err := os.RemoveAll(s.paths.PluginsPath + plugin.Name()); err != nil {
				return err
			}
			continue
		}

		pluginFileName := strings.Split(plugin.Name(), "-")
		pluginName := pluginFileName[0]
		pluginVersion := strings.ReplaceAll(pluginFileName[1], ".jar", "")

		pluginsArr = append(pluginsArr, models.Plugin{
			Name:        pluginName,
			Versions:    nil,
			LastVersion: pluginVersion,
		})
	}

	if err := s.deleteOldAndWrongPlugins(pluginsArr, pluginsInfo); err != nil {
		return err
	}

	if err := s.addNewPlugins(pluginsInfo); err != nil {
		return err
	}

	return nil
}

func (s *PluginService) deleteOldAndWrongPlugins(plugins, pluginsInfo []models.Plugin) error {
	for _, plugin := range plugins {
		var hasPlugin bool
		for _, pluginInfo := range pluginsInfo {
			if plugin.Name == pluginInfo.Name && plugin.LastVersion == pluginInfo.LastVersion {
				hasPlugin = true
			}
		}

		if !hasPlugin {
			if err := os.RemoveAll(s.paths.PluginsPath + plugin.Name + "-" + plugin.LastVersion + ".jar"); err != nil {
				return err
			}
		}
	}

	return nil
}

func (s *PluginService) addNewPlugins(pluginsInfo []models.Plugin) error {
	for _, pluginInfo := range pluginsInfo {
		_, err := os.Stat(s.paths.PluginsPath + pluginInfo.Name + "-" + pluginInfo.LastVersion + ".jar")
		if os.IsNotExist(err) {
			pluginFileBytes, err := s.DownloadPlugin(pluginInfo.Name, pluginInfo.LastVersion)
			if err != nil {
				return err
			}

			if err := s.UpdatePlugin(pluginInfo.Name, pluginInfo.LastVersion, *pluginFileBytes); err != nil {
				return err
			}
		}
	}

	return nil
}

func (s *PluginService) GetPluginsInfo() ([]models.Plugin, error) {
	resp, err := http.Get("https://api.mineleague.ru/plugin")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var response models.PluginsResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, err
	}

	if !response.Success {
		return nil, errors.New("error getting plugins list from API")
	}

	for index, plugin := range response.Plugins {
		versions := make([]*version.Version, len(plugin.Versions))
		for i, raw := range plugin.Versions {
			v, _ := version.NewVersion(raw)
			versions[i] = v
		}
		sort.Sort(version.Collection(versions))
		plugin.LastVersion = versions[len(versions)-1].String()
		response.Plugins[index] = plugin
	}

	return response.Plugins, nil
}

func (s *PluginService) DownloadPlugin(pluginName, version string) (*[]byte, error) {
	resp, err := http.Get("https://api.mineleague.ru/plugin/" + pluginName + "/" + version)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, errors.New("error downloading plugin from API")
	}

	fileBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return &fileBytes, nil
}

func (s *PluginService) UpdatePlugin(pluginName, version string, pluginFileBytes []byte) error {
	file, err := os.Create(s.paths.PluginsPath + pluginName + "-" + version + ".jar")
	if err != nil {
		return err
	}
	defer file.Close()

	if _, err = file.Write(pluginFileBytes); err != nil {
		return err
	}

	return nil
}
