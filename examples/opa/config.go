package main

import (
	"os"

	"gopkg.in/yaml.v3"
)

type Type int

const (
	IntType Type = iota
	StringType
)

type FunctionDeclaration struct {
	Name       string `yaml:"name"`
	Raw        bool   `yaml:"bool"`
	Parameters []Type `yaml:"parameters"`
	Returns    []Type `yaml:"returns"`
}

type Config struct {
	Functions []FunctionDeclaration `yaml:"functions"`
}

func LoadConfig(filePath string) (*Config, error) {
	b, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}
	var cfg Config
	if err := yaml.Unmarshal(b, &cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}
