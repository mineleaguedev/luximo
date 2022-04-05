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

type PaperService struct {
	paths models.Paths
}

func NewPaperService(paths models.Paths) *PaperService {
	return &PaperService{paths: paths}
}

func (s *PaperService) UpdatePaper() error {
	paperInfo, err := s.GetPaperVersionsInfo()
	if err != nil {
		return err
	}

	paperVersions, err := os.ReadDir(s.paths.PaperPath)
	if err != nil {
		return err
	}

	if len(paperVersions) != 1 {
		for _, paperVersion := range paperVersions {
			if err := os.RemoveAll(s.paths.PaperPath + paperVersion.Name()); err != nil {
				return err
			}
		}

		paperFileBytes, err := s.DownloadPaper(paperInfo.LastVersion)
		if err != nil {
			return err
		}

		if err := s.UpdatePaperVersion(paperInfo.LastVersion, *paperFileBytes); err != nil {
			return err
		}

		return nil
	}

	for _, paperVersion := range paperVersions {
		if !strings.Contains(paperVersion.Name(), ".rar") {
			if err := os.RemoveAll(s.paths.PaperPath + paperVersion.Name()); err != nil {
				return err
			}
			continue
		}

		paperFileName := strings.Split(paperVersion.Name(), "-")
		paperVersion := strings.ReplaceAll(paperFileName[1], ".rar", "")

		if paperVersion != paperInfo.LastVersion {
			paperFileBytes, err := s.DownloadPaper(paperInfo.LastVersion)
			if err != nil {
				return err
			}

			if err := s.UpdatePaperVersion(paperInfo.LastVersion, *paperFileBytes); err != nil {
				return err
			}

			return nil
		}
	}

	paperFileBytes, err := s.DownloadPaper(paperInfo.LastVersion)
	if err != nil {
		return err
	}

	if err := s.UpdatePaperVersion(paperInfo.LastVersion, *paperFileBytes); err != nil {
		return err
	}

	return nil
}

func (s *PaperService) GetPaperVersionsInfo() (*models.PaperResponse, error) {
	resp, err := http.Get("https://api.mineleague.ru/paper")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var response models.PaperResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, err
	}

	if !response.Success {
		return nil, errors.New("error getting paper versions list from API")
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

func (s *PaperService) DownloadPaper(version string) (*[]byte, error) {
	resp, err := http.Get("https://api.mineleague.ru/paper/" + version)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, errors.New("error downloading paper from API")
	}

	fileBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return &fileBytes, nil
}

func (s *PaperService) UpdatePaperVersion(version string, paperFileBytes []byte) error {
	file, err := os.Create(s.paths.PaperPath + "paper-" + version + ".rar")
	if err != nil {
		return err
	}
	defer file.Close()

	if _, err = file.Write(paperFileBytes); err != nil {
		return err
	}

	return nil
}
