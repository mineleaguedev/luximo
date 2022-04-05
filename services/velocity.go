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

type VelocityService struct {
	paths models.Paths
}

func NewVelocityService(paths models.Paths) *VelocityService {
	return &VelocityService{paths: paths}
}

func (s *VelocityService) UpdateVelocity() error {
	velocityInfo, err := s.GetVelocityVersionsInfo()
	if err != nil {
		return err
	}

	velocityVersions, err := os.ReadDir(s.paths.VelocityPath)
	if err != nil {
		return err
	}

	if len(velocityVersions) != 1 {
		for _, velocityVersion := range velocityVersions {
			if err := os.RemoveAll(s.paths.VelocityPath + velocityVersion.Name()); err != nil {
				return err
			}
		}

		velocityFileBytes, err := s.DownloadVelocity(velocityInfo.LastVersion)
		if err != nil {
			return err
		}

		if err := s.UpdateVelocityVersion(velocityInfo.LastVersion, *velocityFileBytes); err != nil {
			return err
		}

		return nil
	}

	for _, velocityVersion := range velocityVersions {
		if !strings.Contains(velocityVersion.Name(), ".rar") {
			if err := os.RemoveAll(s.paths.VelocityPath + velocityVersion.Name()); err != nil {
				return err
			}
			continue
		}

		velocityFileName := strings.Split(velocityVersion.Name(), "-")
		velocityVersion := strings.ReplaceAll(velocityFileName[1], ".rar", "")

		if velocityVersion != velocityInfo.LastVersion {
			velocityFileBytes, err := s.DownloadVelocity(velocityInfo.LastVersion)
			if err != nil {
				return err
			}

			if err := s.UpdateVelocityVersion(velocityInfo.LastVersion, *velocityFileBytes); err != nil {
				return err
			}

			return nil
		}
	}

	velocityFileBytes, err := s.DownloadVelocity(velocityInfo.LastVersion)
	if err != nil {
		return err
	}

	if err := s.UpdateVelocityVersion(velocityInfo.LastVersion, *velocityFileBytes); err != nil {
		return err
	}

	return nil
}

func (s *VelocityService) GetVelocityVersionsInfo() (*models.VelocityResponse, error) {
	resp, err := http.Get("https://api.mineleague.ru/velocity")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var response models.VelocityResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, err
	}

	if !response.Success {
		return nil, errors.New("error getting velocity versions list from API")
	}

	versions := make([]*version.Version, len(response.Versions))
	for i, raw := range response.Versions {
		v, _ := version.NewVersion(raw)
		versions[i] = v
	}
	sort.Sort(version.Collection(versions))
	response.LastVersion = versions[len(versions)-1].String()

	return &response, nil
}

func (s *VelocityService) DownloadVelocity(version string) (*[]byte, error) {
	resp, err := http.Get("https://api.mineleague.ru/velocity/" + version)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, errors.New("error downloading velocity from API")
	}

	fileBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return &fileBytes, nil
}

func (s *VelocityService) UpdateVelocityVersion(version string, velocityFileBytes []byte) error {
	file, err := os.Create(s.paths.VelocityPath + "velocity-" + version + ".rar")
	if err != nil {
		return err
	}
	defer file.Close()

	if _, err = file.Write(velocityFileBytes); err != nil {
		return err
	}

	return nil
}
