package main

import (
	"testing"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
)

const testConfigFile = `
labels:
- name: size/xs
  color: 00ff00
  min-lines: 0
  description: "Less than 10 lines"
- name: size/s
  color: 00ff11
  min-lines: 10
  description: "Less than 100 lines"
- name: size/m
  color: 00ff22
  min-lines: 100
  description: "Less than 1000 lines"
`

func TestParsesConfig(t *testing.T) {
	fs = afero.Afero{Fs: afero.NewMemMapFs()}

	err := fs.WriteFile(".github/pr-size-labeler.yml", []byte(testConfigFile), 0644)
	assert.NoError(t, err)

	config, err := loadConfig(".github/pr-size-labeler.yml")
	assert.NoError(t, err)

	expectedLabels := []Label{
		{
			Name:        "size/xs",
			Color:       "00ff00",
			MinLines:    0,
			Description: "Less than 10 lines",
		},
		{
			Name:        "size/s",
			Color:       "00ff11",
			MinLines:    10,
			Description: "Less than 100 lines",
		},
		{
			Name:        "size/m",
			Color:       "00ff22",
			MinLines:    100,
			Description: "Less than 1000 lines",
		},
	}
	assert.Equal(t, expectedLabels, config.Labels)
}
