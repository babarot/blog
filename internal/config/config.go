package config

import (
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"reflect"
	"strings"

	"github.com/MakeNowJust/heredoc"
	"github.com/babarot/blog/internal/blog"
	"github.com/babarot/blog/internal/env"
	"github.com/babarot/blog/internal/shell"
	"github.com/go-playground/validator"
	"github.com/muesli/reflow/indent"
	"gopkg.in/yaml.v2"
)

type KeyType string

const Key KeyType = "config"

type Config struct {
	Posts     []blog.Article `yaml:"-"`
	LogWriter io.Writer      `yaml:"-"`

	Editor string `yaml:"editor"`
	Hugo   Hugo   `yaml:"hugo"`
}

var validate *validator.Validate

type configError struct {
	configPath string
	configDir  string
	parser     parser
	err        error
}

type parser struct{}

type Hugo struct {
	Command    string `yaml:"command"`
	RootDir    string `yaml:"root_dir"`
	ContentDir string `yaml:"content_dir"`
}

func (p parser) getDefaultConfig() Config {
	return Config{
		Editor: "vim",
		Hugo: Hugo{
			Command: "hugo server",
		},
	}
}

func (p parser) getDefaultConfigContents() string {
	defaultConfig := p.getDefaultConfig()
	content, _ := yaml.Marshal(defaultConfig)
	return string(content)
}

func (e configError) Error() string {
	return heredoc.Docf(`
		Couldn't find the "%s" config file.
		Please try again after creating it or specifying a valid config path.
		The recommended config path is %s (default).
		Example YAML file contents:
		---
		%s
		---
		Original error:
		%s
		`,
		e.configPath,
		env.BLOG_CONFIG_PATH,
		e.parser.getDefaultConfigContents(),
		indent.String(e.err.Error(), 2),
	)
}

func (p parser) createConfigFile(path string) error {
	// Ensure directory exists
	if err := p.ensureDirExists(filepath.Dir(path)); err != nil {
		return err
	}

	// Create the config file if missing
	if _, err := os.Stat(path); os.IsNotExist(err) {
		slog.Warn("creating config file as it does not exist", "config-file", path)
		newConfigFile, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_EXCL, 0666)
		if err != nil {
			return err
		}
		defer newConfigFile.Close()

		// Write default config contents
		if err := p.writeConfigFileContents(newConfigFile); err != nil {
			return err
		}
	}

	return nil
}

func (p parser) ensureDirExists(dirPath string) error {
	if _, err := os.Stat(dirPath); os.IsNotExist(err) {
		slog.Warn("creating directory as it does not exist", "dir", dirPath)
		if err := os.MkdirAll(dirPath, os.ModePerm); err != nil {
			return err
		}
	}
	return nil
}

func (p parser) writeConfigFileContents(file *os.File) error {
	_, err := file.WriteString(p.getDefaultConfigContents())
	return err
}

func (p parser) ensureConfigFile() (string, error) {
	path := env.BLOG_CONFIG_PATH

	// Ensure directory exists before creating file
	if err := p.ensureDirExists(filepath.Dir(path)); err != nil {
		return "", err
	}

	// Create file if missing
	if err := p.createConfigFile(path); err != nil {
		return "", configError{
			parser:    p,
			configDir: filepath.Dir(path),
			err:       err,
		}
	}

	return path, nil
}

type parsingError struct {
	err error
}

func (e parsingError) Error() string {
	return fmt.Sprintf("failed to parse config: %v", e.err)
}

func (p parser) readConfigFile(path string) (Config, error) {
	var cfg Config
	data, err := os.ReadFile(path)
	if err != nil {
		return cfg, configError{
			configPath: path,
			configDir:  filepath.Dir(path),
			parser:     p,
			err:        err,
		}
	}

	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return cfg, err
	}

	if err := validate.Struct(cfg); err != nil {
		for _, err := range err.(validator.ValidationErrors) {
			return cfg, fmt.Errorf("validation error: Field %s, %q is invalid\n", err.Field(), err.Value())
		}
	}
	return cfg, nil
}

func initParser() parser {
	validate = validator.New()
	validate.RegisterTagNameFunc(func(fld reflect.StructField) string {
		name := strings.Split(fld.Tag.Get("yaml"), ",")[0]
		if name == "-" {
			return ""
		}
		return name
	})

	return parser{}
}

func Parse(path string) (Config, error) {
	parser := initParser()

	var cfg Config
	var err error
	var configPath string

	if path == "" {
		configPath, err = parser.ensureConfigFile()
		if err != nil {
			return cfg, parsingError{err: err}
		}
	} else {
		configPath = path
	}
	slog.Debug("config file found", "config-file", configPath)

	cfg, err = parser.readConfigFile(configPath)
	if err != nil {
		return cfg, parsingError{err: err}
	}

	rootDir, err := shell.ExpandHome(cfg.Hugo.RootDir)
	if err != nil {
		return cfg, parsingError{err: err}
	}
	cfg.Hugo.RootDir = rootDir

	contentDir, err := shell.ExpandHome(cfg.Hugo.ContentDir)
	if err != nil {
		return cfg, parsingError{err: err}
	}
	cfg.Hugo.ContentDir = contentDir

	return cfg, nil
}
