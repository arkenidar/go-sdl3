// storage.go
package main

import (
	"encoding/json"
	"os"
)

// Item is a single to-do style record persisted to disk.
type Item struct {
	ID   int    `json:"id"`
	Text string `json:"text"`
	Done bool   `json:"done"`
}

// LoadItems reads items from path. A missing file is treated as an empty list.
func LoadItems(path string) ([]Item, error) {
	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return []Item{}, nil
	}
	if err != nil {
		return nil, err
	}

	var items []Item
	if err := json.Unmarshal(data, &items); err != nil {
		return nil, err
	}
	return items, nil
}

// SaveItems writes items to path as indented JSON.
func SaveItems(path string, items []Item) error {
	data, err := json.MarshalIndent(items, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}
