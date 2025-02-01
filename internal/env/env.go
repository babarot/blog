package env

import (
	"os"
	"path/filepath"
)

const (
	defaultXDGConfigDirname = ".config"
	defaultXDGDataDirname   = ".local/share"
)

var (
	BLOG_LOG_PATH    string
	BLOG_CONFIG_PATH string
)

func init() {
	// https://github.com/charmbracelet/log/issues/35
	os.Setenv("CLICOLOR_FORCE", "1")

	// Follow https://specifications.freedesktop.org/basedir-spec/latest/
	if e := os.Getenv("BLOG_CONFIG_PATH"); e == "" {
		configDir := os.Getenv("XDG_CONFIG_HOME")
		if configDir == "" {
			homeDir, err := os.UserHomeDir()
			if err != nil {
				panic(err)
			}
			configDir = filepath.Join(homeDir, defaultXDGConfigDirname)
		}
		BLOG_CONFIG_PATH = filepath.Join(configDir, "blog", "config.yaml")
	}

	if e := os.Getenv("BLOG_LOG_PATH"); e == "" {
		dataDir := os.Getenv("XDG_DATA_HOME")
		if dataDir == "" {
			homeDir, err := os.UserHomeDir()
			if err != nil {
				panic(err)
			}
			dataDir = filepath.Join(homeDir, defaultXDGDataDirname)
		}
		BLOG_LOG_PATH = filepath.Join(dataDir, "blog", "debug.log")
	}
}
