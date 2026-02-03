package quest

import (
	"fmt"
	"path/filepath"
)

func DefaultScriptPath(dataPath string) string {
	return filepath.Join(dataPath, "quest_injector.js")
}

func EnsureScript(cfg Config, dataPath string) (string, error) {
	if !cfg.Enabled {
		return "", nil
	}
	scriptPath := DefaultScriptPath(dataPath)
	if err := WriteAutomatorScript(scriptPath); err != nil {
		return "", fmt.Errorf("write quest script: %w", err)
	}
	return scriptPath, nil
}
