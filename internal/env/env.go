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
	BLOG_LOG_PATH string
)

func init() {
	// https://github.com/charmbracelet/log/issues/35
	os.Setenv("CLICOLOR_FORCE", "1")

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
