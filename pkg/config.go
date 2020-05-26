package pkg

import (
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
)

// Project defines the struct representation of rubik.toml
type Project struct {
	Name         string `toml:"name"`
	Path         string `toml:"path"`
	Watchable    bool   `toml:"watch"`
	Communicable bool   `toml:"communicate"`
}

// Config is the main config for your rubik runtime
// this is declared inside a rubik.toml file
type Config struct {
	ProjectName string `toml:"name"`
	Module      string `toml:"module"`
	IsFlat      bool   `toml:"flat"`
	MaxProcs    int    `toml:"maxprocs"`
	Log         bool
	App         []Project                    `toml:"app"`
	X           map[string]map[string]string `toml:"x"`
}

var sep = string(os.PathSeparator)

// GetTemplateFolderPath returns the absolute template dir path
func GetTemplateFolderPath() string {
	dir, _ := os.Getwd()
	return dir + sep + "templates"
}

// GetStaticFolderPath returns the absolute static dir path
func GetStaticFolderPath() string {
	return filepath.Join(".", "static")
}

// GetRubikConfigPath returns path of rubik config of current project
func GetRubikConfigPath() string {
	dir, _ := os.Getwd()
	return dir + sep + "rubik.toml"
}

// GetRubikConfig returns cherry config
func GetRubikConfig() *Config {
	configPath := GetRubikConfigPath()
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return &Config{}
	}

	var config Config
	_, err := toml.DecodeFile(configPath, &config)
	if err != nil {
		WarnMsg("rubik.toml was found but could not parse it. Error: " + err.Error())
		return &Config{}
	}
	return &config
}

// MakeAndGetCacheDirPath returns rubik's cache dir
func MakeAndGetCacheDirPath() string {
	pwd, _ := os.UserHomeDir()
	path := pwd + sep + ".rubik"
	if _, err := os.Stat(path); os.IsNotExist(err) {
		os.MkdirAll(path, 0755)
	}
	return path
}

// GetErrorHTMLPath ...
func GetErrorHTMLPath() string {
	cacheFolder := filepath.Join(MakeAndGetCacheDirPath(), "cache")
	os.MkdirAll(cacheFolder, 0755)
	return filepath.Join(cacheFolder, "error.html")
}

// OverrideValues writes over the source map with env map
func OverrideValues(source, env map[string]interface{}) map[string]interface{} {
	for k, v := range env {
		source[k] = v
	}
	return source
}
