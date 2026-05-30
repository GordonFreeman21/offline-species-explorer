package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type App struct{}

func (a *App) SearchSpecies(query string) ([]Species, error) {
	if query = strings.TrimSpace(query); query == "" {
		return []Species{}, nil
	}
	searchFunc := func(q string) ([]Species, error) {
		stmt, err := db.Prepare(`
			SELECT id, common_name, scientific_name, kingdom, phylum, class, "order", family, last_synced
			FROM species WHERE common_name LIKE ? OR scientific_name LIKE ?
			ORDER BY common_name LIMIT 100
		`)
		if err != nil {
			return nil, fmt.Errorf("search prepare: %w", err)
		}
		defer stmt.Close()
		pattern := "%" + q + "%"
		rows, err := stmt.Query(pattern, pattern)
		if err != nil {
			return nil, fmt.Errorf("search query: %w", err)
		}
		defer rows.Close()
		var results []Species
		for rows.Next() {
			var s Species
			err := rows.Scan(&s.ID, &s.CommonName, &s.ScientificName, &s.Kingdom, &s.Phylum, &s.Class, &s.Order, &s.Family, &s.LastSynced)
			if err != nil {
				return nil, fmt.Errorf("search scan: %w", err)
			}
			results = append(results, s)
		}
		return results, rows.Err()
	}
	return searchFunc(query)
}

func (a *App) GetSpeciesByID(id int) (Species, error) {
	stmt, err := db.Prepare(`SELECT id, common_name, scientific_name, kingdom, phylum, class, "order", family, last_synced FROM species WHERE id = ?`)
	if err != nil {
		return Species{}, fmt.Errorf("prepare: %w", err)
	}
	defer stmt.Close()
	var s Species
	err = stmt.QueryRow(id).Scan(&s.ID, &s.CommonName, &s.ScientificName, &s.Kingdom, &s.Phylum, &s.Class, &s.Order, &s.Family, &s.LastSynced)
	if err != nil {
		return Species{}, fmt.Errorf("query: %w", err)
	}
	return s, nil
}

func (a *App) SyncSpecies(scientificName string) (Species, error) {
	species, err := fetchFromAPI(scientificName)
	if err != nil {
		return Species{}, fmt.Errorf("API fetch failed: %w", err)
	}
	if err := UpdateSpecies(species); err != nil {
		return Species{}, fmt.Errorf("local update failed: %w", err)
	}
	updated, err := getSpeciesByScientificName(species.ScientificName)
	if err != nil {
		return species, nil
	}
	return updated, nil
}

func (a *App) FetchAndSaveFromAPI(name string) (Species, error) {
	species, err := fetchFromAPI(name)
	if err != nil {
		return Species{}, fmt.Errorf("API fetch failed: %w", err)
	}
	existingID, exists, err := speciesExists(species.ScientificName)
	if err != nil {
		return Species{}, fmt.Errorf("check existing: %w", err)
	}
	if exists {
		species.ID = existingID
		if err := UpdateSpecies(species); err != nil {
			return Species{}, fmt.Errorf("update failed: %w", err)
		}
		return species, nil
	}
	newID, err := InsertSpecies(species)
	if err != nil {
		return Species{}, fmt.Errorf("insert failed: %w", err)
	}
	species.ID = int(newID)
	return species, nil
}

func fetchFromAPI(name string) (Species, error) {
	apiURL := fmt.Sprintf("https://nimbus-api-gxuc.onrender.com/api/v1/bio/species?name=%s", url.QueryEscape(name))
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Get(apiURL)
	if err != nil {
		return Species{}, fmt.Errorf("HTTP request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(io.LimitReader(resp.Body, 65536))
	if err != nil {
		return Species{}, fmt.Errorf("reading response: %w", err)
	}

	contentType := resp.Header.Get("Content-Type")
	if !strings.Contains(contentType, "application/json") {
		snippet := string(body)
		if len(snippet) > 500 {
			snippet = snippet[:500]
		}
		return Species{}, fmt.Errorf("API returned non-JSON (status %d, type=%q): %s", resp.StatusCode, contentType, snippet)
	}

	if resp.StatusCode == http.StatusNotFound {
		return Species{}, fmt.Errorf("species not found in API")
	}
	if resp.StatusCode != http.StatusOK {
		return Species{}, fmt.Errorf("API returned status %d: %s", resp.StatusCode, truncateString(string(body), 300))
	}

	var s Species
	if err := json.Unmarshal(body, &s); err != nil {
		return Species{}, fmt.Errorf("parsing JSON: %w", err)
	}
	if s.ScientificName == "" {
		return Species{}, fmt.Errorf("invalid API response: missing scientific_name")
	}
	return s, nil
}

func truncateString(s string, n int) string {
	if len(s) > n {
		return s[:n] + "..."
	}
	return s
}

func getSpeciesByScientificName(name string) (Species, error) {
	stmt, err := db.Prepare(`SELECT id, common_name, scientific_name, kingdom, phylum, class, "order", family, last_synced FROM species WHERE scientific_name = ?`)
	if err != nil {
		return Species{}, err
	}
	defer stmt.Close()
	var s Species
	err = stmt.QueryRow(name).Scan(&s.ID, &s.CommonName, &s.ScientificName, &s.Kingdom, &s.Phylum, &s.Class, &s.Order, &s.Family, &s.LastSynced)
	return s, err
}

func speciesExists(scientificName string) (int, bool, error) {
	var id int
	err := db.QueryRow(`SELECT id FROM species WHERE scientific_name = ?`, scientificName).Scan(&id)
	if err != nil {
		return 0, false, nil
	}
	return id, true, nil
}
