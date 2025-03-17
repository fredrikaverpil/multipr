package config

import (
	"os"

	"gopkg.in/yaml.v3"
)

// JobConfig represents the YAML job configuration.
type JobConfig struct {
	Name string `yaml:"name"`

	Search struct {
		GitHub struct {
			Method string `yaml:"method"`
			Query  string `yaml:"query"`
		} `yaml:"github"`
	} `yaml:"search"`

	Identify []struct {
		Name string `yaml:"name"`
		Cmd  string `yaml:"cmd"`
	} `yaml:"identify"`

	Changes []struct {
		Name string `yaml:"name"`
		Cmd  string `yaml:"cmd"`
	} `yaml:"changes"`

	PR struct {
		GitHub struct {
			Title  string `yaml:"title"`
			Body   string `yaml:"body"`
			Branch string `yaml:"branch"`
		} `yaml:"github"`
	} `yaml:"pr"`
}

// LoadFromFile loads a Config from a YAML file path.
func LoadFromFile(path string) (*JobConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var cfg JobConfig
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}
