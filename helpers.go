package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/portapps/discord-ptb-portable/assets"
)

const defaultSettings = `{
  "SKIP_HOST_UPDATE": true,
  "USE_PINNED_UPDATE_MANIFEST": true
}`

func ensureSettings(settingsPath string) error {
	if _, err := os.Stat(settingsPath); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return writeSettingsJSON(settingsPath, []byte(defaultSettings))
		}
		return fmt.Errorf("stat settings.json: %w", err)
	}

	rawSettings, err := os.ReadFile(settingsPath)
	if err != nil {
		return fmt.Errorf("read settings.json: %w", err)
	}

	jsonMapSettings := make(map[string]interface{})
	if err = json.Unmarshal(rawSettings, &jsonMapSettings); err != nil {
		backupPath := settingsPath + ".bak"
		if renameErr := os.Rename(settingsPath, backupPath); renameErr != nil {
			return fmt.Errorf("backup invalid settings.json: %w", renameErr)
		}
		return writeSettingsJSON(settingsPath, []byte(defaultSettings))
	}

	jsonMapSettings["SKIP_HOST_UPDATE"] = true
	jsonMapSettings["USE_PINNED_UPDATE_MANIFEST"] = true

	jsonSettings, err := json.MarshalIndent(jsonMapSettings, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal settings.json: %w", err)
	}

	return writeSettingsJSON(settingsPath, jsonSettings)
}

func writeSettingsJSON(settingsPath string, settings []byte) error {
	return os.WriteFile(settingsPath, settings, 0644)
}

func writeAssetFile(assetName, destination string) error {
	assetData, err := assets.Asset(assetName)
	if err != nil {
		return fmt.Errorf("load asset %s: %w", assetName, err)
	}
	if err := os.MkdirAll(filepath.Dir(destination), 0755); err != nil {
		return fmt.Errorf("create destination dir: %w", err)
	}
	return os.WriteFile(destination, assetData, 0644)
}
