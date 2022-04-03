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
	pluginsVersions := map[string]string{}

	plugins, err := os.ReadDir(s.paths.PluginsPath)
	if err != nil {
		return err
	}

	if len(plugins) == 0 {
		return s.update(pluginsVersions)
	}

	var pluginsToDelete []string

	for _, plugin := range plugins {
		pluginFileName := strings.Split(plugin.Name(), "-")
		pluginName := pluginFileName[0]

		if _, ok := pluginsVersions[pluginName]; ok {
			pluginsToDelete = append(pluginsToDelete, pluginName)
			continue
		}

		pluginVersion := strings.ReplaceAll(pluginFileName[1], ".jar", "")
		pluginsVersions[pluginName] = pluginVersion
	}

	for _, pluginToDelete := range pluginsToDelete {
		delete(pluginsVersions, pluginToDelete)
	}

	return s.update(pluginsVersions)
}

func (s *PluginService) update(plugins map[string]string) error {
	pluginsInfo, err := s.GetPluginsInfo()
	if err != nil {
		return err
	}

	for _, pluginInfo := range pluginsInfo {
		var pluginVersion *string
		for pluginName, pVersion := range plugins {
			if strings.Contains(pluginName, pluginInfo.Name) {
				ver := strings.ReplaceAll(pVersion, ".jar", "")
				pluginVersion = &ver
				break
			}
		}

		if pluginVersion != nil && pluginInfo.LastVersion == *pluginVersion {
			continue
		}

		pluginFileBytes, err := s.DownloadPlugin(pluginInfo.Name, pluginInfo.LastVersion)
		if err != nil {
			return err
		}

		if err := s.UpdatePlugin(pluginInfo.Name, pluginInfo.LastVersion, *pluginFileBytes); err != nil {
			return err
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
	if err := s.DeletePlugin(pluginName); err != nil {
		return err
	}

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

func (s *PluginService) DeletePlugin(pluginName string) error {
	plugins, err := os.ReadDir(s.paths.PluginsPath)
	if err != nil {
		return err
	}

	for _, plugin := range plugins {
		if !strings.Contains(plugin.Name(), pluginName) {
			continue
		}

		if err := os.RemoveAll(s.paths.PluginsPath + plugin.Name()); err != nil {
			return err
		}
	}

	return nil
}
