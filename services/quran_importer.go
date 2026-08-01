package services

import (
	"encoding/json"
	"io"
	"net/http"
	"os"
	"strings"
)

type QuranImporter struct {
	apiURL string
}

func NewQuranImporter() *QuranImporter {
	return &QuranImporter{
		apiURL: "http://api.alquran.cloud/v1/quran/quran-simple",
	}
}

type SurahData struct {
	Number int
	Name   string
	Ayahs  []string
}

// FetchAllSurahs returns all Surahs with their ayahs.
func (q *QuranImporter) FetchAllSurahs() ([]SurahData, error) {
	cacheFile := "quran_cache.json"
	var bodyBytes []byte

	// Try to read from local cache first
	if fileInfo, err := os.Stat(cacheFile); err == nil && !fileInfo.IsDir() {
		bodyBytes, err = os.ReadFile(cacheFile)
		if err != nil {
			return nil, err
		}
	} else {
		// If cache doesn't exist, fetch from API
		resp, err := http.Get(q.apiURL)
		if err != nil {
			return nil, err
		}
		defer resp.Body.Close()

		bodyBytes, err = io.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}

		// Save to cache for future uses
		_ = os.WriteFile(cacheFile, bodyBytes, 0644)
	}

	var result struct {
		Data struct {
			Surahs []struct {
				Number int    `json:"number"`
				Name   string `json:"englishName"` // Using englishName for readability (e.g. Al-Fatiha)
				Ayahs  []struct {
					Text string `json:"text"`
				} `json:"ayahs"`
			} `json:"surahs"`
		} `json:"data"`
	}

	if err := json.Unmarshal(bodyBytes, &result); err != nil {
		return nil, err
	}

	var allSurahs []SurahData
	for _, surah := range result.Data.Surahs {
		sd := SurahData{
			Number: surah.Number,
			Name:   surah.Name,
		}
		for _, ayah := range surah.Ayahs {
			text := strings.TrimSpace(ayah.Text)
			if text != "" {
				sd.Ayahs = append(sd.Ayahs, text)
			}
		}
		allSurahs = append(allSurahs, sd)
	}
	return allSurahs, nil
}
