package main

import (
	"os"

	"github.com/portapps/discord-ptb-portable/quest"
)

type QuestConfig struct {
	Enabled bool `yaml:"enabled" mapstructure:"enabled"`
}

func applyQuestScript(cfg QuestConfig, dataPath string) error {
	if !cfg.Enabled {
		return nil
	}
	scriptPath, err := quest.EnsureScript(quest.Config{Enabled: true}, dataPath)
	if err != nil {
		return err
	}
	if scriptPath != "" {
		return os.Setenv("DISCORD_PTB_QUEST_SCRIPT", scriptPath)
	}
	return nil
}
