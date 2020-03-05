package cherry

type ProjectConfig struct {
	Name string `toml:"name"`
}

type BaseConfig struct {
	Port string `toml:"port"`
}

type Environment struct {
	Default     BaseConfig `toml:"default"`
	Development BaseConfig `toml:"development"`
	Production  BaseConfig `toml:"production"`
}

type Config struct {
	Project ProjectConfig `toml:"project"`
	Cherry  Environment   `toml:"cherry"`
}
