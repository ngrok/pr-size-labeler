package main

import (
	"github.com/google/go-github/v50/github"
	"github.com/spf13/afero"
	"gopkg.in/yaml.v3"
)

var fs = afero.Afero{Fs: afero.NewOsFs()}

type Label struct {
	Name        string `yaml:"name"`
	Color       string `yaml:"color"`
	MinLines    int    `yaml:"min-lines"`
	Description string `yaml:"description"`
}

func (l Label) Matches(label github.Label) bool {
	return label.Name != nil && *label.Name == l.Name &&
		label.Color != nil && *label.Color == l.Color &&
		label.Description != nil && *label.Description == l.Description
}

type Config struct {
	IgnoreLinguistGenerated bool    `yaml:"ignore-linguist-generated"`
	Labels                  []Label `yaml:"labels"`
}

func loadConfig(path string) (Config, error) {
	configFile, err := fs.ReadFile(path)
	if err != nil {
		return Config{}, err
	}

	c := Config{}
	err = yaml.Unmarshal(configFile, &c)
	return c, err
}
