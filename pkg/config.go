package pkg

import (
	"errors"
	"os"
	"strings"

	"github.com/BurntSushi/toml"
)

// NotationMap is a type of map structure that can get you the value of a
// embedded key inside a map
type NotationMap map[string]interface{}

// Get values of key using dot notations from NotationMap
func (nm NotationMap) Get(accessor string) (interface{}, error) {
	var fields = []string{}

	if strings.Contains(accessor, ".") {
		fields = strings.Split(accessor, ".")
	} else {
		fields = append(fields, accessor)
	}

	// then we have entered the arena of dot notations
	if len(fields) > 1 {
		var finalValue interface{}
		for _, f := range fields {
			if finalValue == nil {
				if nm[f] == nil {
					return nil, noSuchKeyErr(f, accessor)
				}
				finalValue = nm[f]
			} else {
				interm, ok := finalValue.(map[string]interface{})
				if !ok {
					return nil, noSuchKeyErr(f, accessor)
				}
				if interm[f] == nil {
					return nil, noSuchKeyErr(f, accessor)
				}
				finalValue = interm[f]
			}
		}
		return finalValue, nil
	}

	return nm[accessor], nil
}

// Set value of a accessor using dot notations from NotationMap
func (nm NotationMap) Set(accessor string, value interface{}) error {
	var fields = []string{}

	if strings.Contains(accessor, ".") {
		fields = strings.Split(accessor, ".")
	} else {
		fields = append(fields, accessor)
	}

	// then we have entered the arena of dot notations
	if len(fields) > 1 {
		var finalValue interface{}
		for i, f := range fields {
			if finalValue == nil {
				if nm[f] == nil {
					return noSuchKeyErr(f, accessor)
				}

				// then this is the last key that we need to traverse to
				if i == len(fields) {
					nm[f] = value
				}
			} else {
				interm, ok := finalValue.(map[string]interface{})
				if !ok {
					return noSuchKeyErr(f, accessor)
				}
				if interm[f] == nil {
					return noSuchKeyErr(f, accessor)
				}
				interm[f] = value
			}
		}
		return nil
	}

	return nil
}

func noSuchKeyErr(key, acc string) error {
	return errors.New("no such key: " + key + " for notation: " + acc)
}

type Project struct {
	Name         string `toml:"name"`
	Path         string `toml:"path"`
	Watchable    bool   `toml:"watchable"`
	Communicable bool   `toml:"communicable"`
}

// Config is the main config for your cherry server
type Config struct {
	ProjectName string    `toml:"name"`
	IsFlat      bool      `toml:"flat"`
	App         []Project `toml:"app"`
}

// GetCherryConfigPath returns path of cherry config of current project
func GetCherryConfigPath() string {
	dir, _ := os.Getwd()
	return dir + string(os.PathSeparator) + "sketch.toml"
}

// GetCherryConfig returns cherry config
func GetCherryConfig() *Config {
	configPath := GetCherryConfigPath()
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		DebugMsg("Did not find sketch.toml. Booting server without one.")
		return &Config{}
	}

	var config Config
	_, err := toml.DecodeFile(configPath, &config)
	if err != nil {
		WarnMsg("sketch.toml was found but could not parse it. Error: " + err.Error())
		return &Config{}
	}
	return &config
}
