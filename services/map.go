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

type MapService struct {
	paths models.Paths
}

func NewMapService(paths models.Paths) *MapService {
	return &MapService{paths: paths}
}

func (s *MapService) UpdateMaps() error {
	var minigamesArr []models.MiniGames

	mapsInfo, err := s.GetMapsInfo()
	if err != nil {
		return err
	}

	minigames, err := os.ReadDir(s.paths.MapsPath)
	if err != nil {
		return err
	}

	if len(minigames) == 0 {
		if err := s.deleteOldAndWrongMaps(minigamesArr, mapsInfo); err != nil {
			return err
		}

		if err := s.addNewMaps(mapsInfo); err != nil {
			return err
		}
	}

	for _, minigame := range minigames {
		formats, err := os.ReadDir(s.paths.MapsPath + minigame.Name())
		if err != nil {
			return err
		}

		var formatsArr []models.Format
		for _, format := range formats {
			maps, err := os.ReadDir(s.paths.MapsPath + minigame.Name() + "/" + format.Name())
			if err != nil {
				return err
			}

			var mapsArr []models.Map
			for _, mapVersion := range maps {
				mapVersionFolderName := strings.Split(mapVersion.Name(), "-")
				mapName := mapVersionFolderName[0]
				mapVersionName := mapVersionFolderName[1]

				mapFiles, err := os.ReadDir(s.paths.MapsPath + minigame.Name() + "/" + format.Name() + "/" + mapVersion.Name())
				if err != nil {
					return err
				}

				var hasWorld bool
				var hasConfig bool
				for _, mapFile := range mapFiles {
					if mapFile.Name() == "world.rar" {
						hasWorld = true
					} else if mapFile.Name() == "map.yml" {
						hasConfig = true
					} else {
						if err := os.RemoveAll(s.paths.MapsPath + minigame.Name() + "/" + format.Name() + "/" + mapVersion.Name() + "/" + mapFile.Name()); err != nil {
							return err
						}
					}
				}

				mapsArr = append(mapsArr, models.Map{
					Name:        mapName,
					LastVersion: mapVersionName,
					HasWorld:    hasWorld,
					HasConfig:   hasConfig,
				})
			}

			formatsArr = append(formatsArr, models.Format{
				Format: format.Name(),
				Maps:   mapsArr,
			})
		}

		minigamesArr = append(minigamesArr, models.MiniGames{
			Name:    minigame.Name(),
			Formats: formatsArr,
		})
	}

	if err := s.deleteOldAndWrongMaps(minigamesArr, mapsInfo); err != nil {
		return err
	}

	if err := s.addNewMaps(mapsInfo); err != nil {
		return err
	}

	return nil
}

func (s *MapService) deleteOldAndWrongMaps(minigamesArr []models.MiniGames, mapsInfo []models.MiniGames) error {
	for _, minigame := range minigamesArr {
		var hasMinigame bool
		for _, minigameInfo := range mapsInfo {
			if minigameInfo.Name != minigame.Name {
				continue
			}
			hasMinigame = true

			for _, format := range minigame.Formats {
				var hasFormat bool
				for _, formatInfo := range minigameInfo.Formats {
					if formatInfo.Format != format.Format {
						continue
					}
					hasFormat = true

					for _, formatMap := range format.Maps {
						var hasMap bool
						for _, mapInfo := range formatInfo.Maps {
							if formatMap.Name != mapInfo.Name {
								continue
							}
							hasMap = true

							if (formatMap.LastVersion != mapInfo.LastVersion) || !formatMap.HasWorld || !formatMap.HasConfig {
								if err := os.RemoveAll(s.paths.MapsPath + minigame.Name + "/" + format.Format + "/" + formatMap.Name + "-" + formatMap.LastVersion); err != nil {
									return err
								}
							}
						}

						if !hasMap {
							if err := os.RemoveAll(s.paths.MapsPath + minigame.Name + "/" + format.Format + "/" + formatMap.Name + "-" + formatMap.LastVersion); err != nil {
								return err
							}
						}
					}
				}

				if !hasFormat {
					if err := os.RemoveAll(s.paths.MapsPath + minigame.Name + "/" + format.Format); err != nil {
						return err
					}
				}
			}
		}

		if !hasMinigame {
			if err := os.RemoveAll(s.paths.MapsPath + minigame.Name); err != nil {
				return err
			}
		}
	}

	return nil
}

func (s *MapService) addNewMaps(mapsInfo []models.MiniGames) error {
	for _, minigameInfo := range mapsInfo {
		for _, formatInfo := range minigameInfo.Formats {
			for _, mapInfo := range formatInfo.Maps {
				_, err := os.Stat(s.paths.MapsPath + minigameInfo.Name + "/" + formatInfo.Format + "/" + mapInfo.Name + "-" + mapInfo.LastVersion)
				if os.IsNotExist(err) {
					mapWorldFileBytes, err := s.DownloadMapWorld(minigameInfo.Name, formatInfo.Format, mapInfo.Name, mapInfo.LastVersion)
					if err != nil {
						return err
					}

					mapConfigFileBytes, err := s.DownloadMapConfig(minigameInfo.Name, formatInfo.Format, mapInfo.Name, mapInfo.LastVersion)
					if err != nil {
						return err
					}

					if err := s.UpdateMap(minigameInfo.Name, formatInfo.Format, mapInfo.Name, mapInfo.LastVersion, mapWorldFileBytes, mapConfigFileBytes); err != nil {
						return err
					}
				}
			}
		}
	}

	return nil
}

func (s *MapService) GetMapsInfo() ([]models.MiniGames, error) {
	resp, err := http.Get("https://api.mineleague.ru/map")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var response models.MapsResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, err
	}

	if !response.Success {
		return nil, errors.New("error getting maps list from API")
	}

	for minigameIndex, minigame := range response.MiniGames {
		for formatIndex, format := range minigame.Formats {
			for mapIndex, minigameMap := range format.Maps {
				versions := make([]*version.Version, len(minigameMap.Versions))
				for i, raw := range minigameMap.Versions {
					v, _ := version.NewVersion(raw)
					versions[i] = v
				}
				sort.Sort(version.Collection(versions))
				minigameMap.LastVersion = versions[len(versions)-1].String()
				response.MiniGames[minigameIndex].Formats[formatIndex].Maps[mapIndex] = minigameMap
			}
		}
	}

	return response.MiniGames, nil
}

func (s *MapService) DownloadMapWorld(minigame, format, minigameMap, version string) (*[]byte, error) {
	resp, err := http.Get("https://api.mineleague.ru/map/" + minigame + "/" + format + "/" + minigameMap + "/" + version + "/world")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, errors.New("error downloading map world from API")
	}

	fileBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return &fileBytes, nil
}

func (s *MapService) DownloadMapConfig(minigame, format, minigameMap, version string) (*[]byte, error) {
	resp, err := http.Get("https://api.mineleague.ru/map/" + minigame + "/" + format + "/" + minigameMap + "/" + version + "/config")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, errors.New("error downloading map config from API")
	}

	fileBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return &fileBytes, nil
}

func (s *MapService) UpdateMap(minigame, format, mapName, version string, mapWorldFileBytes, mapConfigFileBytes *[]byte) error {
	if err := s.DeleteMap(minigame, format, mapName); err != nil {
		return err
	}

	if mapWorldFileBytes != nil {
		if err := os.MkdirAll(s.paths.MapsPath+minigame+"/"+format+"/"+mapName+"-"+version, 0755); err != nil {
			return err
		}

		worldFile, err := os.Create(s.paths.MapsPath + minigame + "/" + format + "/" + mapName + "-" + version + "/" + "world.rar")
		if err != nil {
			return err
		}
		defer worldFile.Close()

		if _, err = worldFile.Write(*mapWorldFileBytes); err != nil {
			return err
		}
	}

	if mapConfigFileBytes != nil {
		if err := os.MkdirAll(s.paths.MapsPath+minigame+"/"+format+"/"+mapName+"-"+version, 0755); err != nil {
			return err
		}

		configFile, err := os.Create(s.paths.MapsPath + minigame + "/" + format + "/" + mapName + "-" + version + "/" + "map.yml")
		if err != nil {
			return err
		}
		defer configFile.Close()

		if _, err = configFile.Write(*mapConfigFileBytes); err != nil {
			return err
		}
	}

	return nil
}

func (s *MapService) DeleteMap(minigameName, formatName, mapName string) error {
	if err := os.RemoveAll(s.paths.MapsPath + minigameName + "/" + formatName + "/" + mapName); err != nil {
		return err
	}

	return nil
}
