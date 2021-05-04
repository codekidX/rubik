package pkg

import (
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
	"github.com/pkg/errors"
)

// LoggingConfig is used by Rubik's Log channels to log out the
// message provided by the developers in a stream mentioned by
// the developer in the rubik.toml [[app.logging]] config
//
// Supported streams are stdout, stderr, file. Path is ignored
// if one of the standard os stream and it is only used for file
// stream.
// Format supports formatting with the following placeholders
// 		- () = Date format, $level and $message
type LoggingConfig struct {
	Stream      string `toml:"stream"`
	ErrorStream string `toml:"err_stream"`
	Path        string `toml:"path"`
	Format      string `toml:"format"`
}

// Project defines the struct representation of rubik.toml
type Project struct {
	Name         string        `toml:"name"`
	Path         string        `toml:"path"`
	Watchable    bool          `toml:"watch"`
	Communicable bool          `toml:"communicate"`
	RunCommand   string        `toml:"run_command"`
	Logging      LoggingConfig `toml:"logging"`
}

// WorkspaceConfig is the main WorkspaceConfig for your rubik runtime
// this is declared inside a rubik.toml file
type WorkspaceConfig struct {
	ProjectName string                       `toml:"name"`
	Module      string                       `toml:"module"`
	IsFlat      bool                         `toml:"flat"`
	MaxProcs    int                          `toml:"maxprocs"`
	App         []Project                    `toml:"app"`
	X           map[string]map[string]string `toml:"x"`
	Pod         map[string]string            `toml:"pod"`
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

// GetRubikConfig returns Config: a structural representation of rubik.toml
func GetRubikConfig() (*WorkspaceConfig, error) {
	configPath := GetRubikConfigPath()
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return nil, errors.New("Not a rubik project")
	}

	return GetWorkspaceConfig(configPath)
}

// GetWorkspaceConfig will give the decoded WorkspaceConfig type
// if it can parse the rubik.toml with the given path
func GetWorkspaceConfig(path string) (*WorkspaceConfig, error) {
	var config WorkspaceConfig
	_, err := toml.DecodeFile(path, &config)
	if err != nil {
		WarnMsg("rubik.toml was found but could not parse it. Error: " + err.Error())
		return nil, errors.New("Cannot parse rubik.toml, please verify if it is valid TOML file")
	}
	return &config, nil
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
