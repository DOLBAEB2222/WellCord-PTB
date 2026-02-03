//go:generate go install -v github.com/kevinburke/go-bindata/v4/go-bindata
//go:generate go-bindata -prefix res/ -pkg assets -o assets/assets.go res/DiscordPTB.lnk res/pinned_update.json
//go:generate go install -v github.com/josephspurrier/goversioninfo/cmd/goversioninfo
//go:generate goversioninfo -icon=res/papp.ico -manifest=res/papp.manifest
package main

import (
	"os"
	"path"

	"github.com/portapps/portapps/v3"
	"github.com/portapps/portapps/v3/pkg/log"
	"github.com/portapps/portapps/v3/pkg/registry"
	"github.com/portapps/portapps/v3/pkg/shortcut"
	"github.com/portapps/portapps/v3/pkg/utl"
)

type config struct {
	Cleanup bool        `yaml:"cleanup" mapstructure:"cleanup"`
	Proxy   ProxyConfig `yaml:"proxy" mapstructure:"proxy"`
	Quest   QuestConfig `yaml:"quest" mapstructure:"quest"`
}

var (
	app *portapps.App
	cfg *config
)

func init() {
	var err error

	// Default config
	cfg = &config{
		Cleanup: false,
		Proxy: ProxyConfig{
			Mode: "system",
		},
		Quest: QuestConfig{
			Enabled: true,
		},
	}

	// Init app
	if app, err = portapps.NewWithCfg("discord-ptb-portable", "DiscordPTB", cfg); err != nil {
		log.Fatal().Err(err).Msg("Cannot initialize application. See log file for more info.")
	}
}

func main() {
	utl.CreateFolder(app.DataPath)
	electronAppPath := app.ElectronAppPath()

	app.Process = utl.PathJoin(electronAppPath, "DiscordPTB.exe")
	app.Args = []string{
		"--user-data-dir=" + app.DataPath,
	}
	app.WorkingDir = electronAppPath
	applyProxyArgs(cfg.Proxy, &app.Args)
	if err := applyQuestScript(cfg.Quest, app.DataPath); err != nil {
		log.Error().Err(err).Msg("Cannot write quest automator script")
	}

	// Cleanup on exit
	if cfg.Cleanup {
		defer func() {
			regKey := registry.Key{
				Key:  `HKCU\SOFTWARE\DiscordPTB`,
				Arch: "32",
			}
			if regKey.Exists() {
				if err := regKey.Delete(true); err != nil {
					log.Error().Err(err).Msg("Cannot remove registry key")
				}
			}
			utl.Cleanup([]string{
				path.Join(os.Getenv("APPDATA"), "discordptb"),
				path.Join(os.Getenv("TEMP"), "Discord Crashes"),
			})
		}()
	}

	// Update settings
	settingsPath := utl.PathJoin(app.DataPath, "settings.json")
	if err := ensureSettings(settingsPath); err != nil {
		log.Error().Err(err).Msg("Cannot update settings.json")
	}

	// Copy pinned_update.json
	if err := writeAssetFile("pinned_update.json", utl.PathJoin(app.DataPath, "pinned_update.json")); err != nil {
		log.Error().Err(err).Msg("Cannot write pinned_update.json")
	}

	// Copy default shortcut
	shortcutPath := path.Join(utl.StartMenuPath(), "Discord PTB Portable.lnk")
	if err := writeAssetFile("DiscordPTB.lnk", shortcutPath); err != nil {
		log.Error().Err(err).Msg("Cannot write default shortcut")
	}

	// Update default shortcut
	err = shortcut.Create(shortcut.Shortcut{
		ShortcutPath:     shortcutPath,
		TargetPath:       app.Process,
		Arguments:        shortcut.Property{Clear: true},
		Description:      shortcut.Property{Value: "Discord PTB Portable by Portapps"},
		IconLocation:     shortcut.Property{Value: app.Process},
		WorkingDirectory: shortcut.Property{Value: app.AppPath},
	})
	if err != nil {
		log.Error().Err(err).Msg("Cannot create shortcut")
	}
	defer func() {
		if err := os.Remove(shortcutPath); err != nil {
			log.Error().Err(err).Msg("Cannot remove shortcut")
		}
	}()

	defer app.Close()
	app.Launch(os.Args[1:])
}
