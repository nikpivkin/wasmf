package main

import (
	"os"

	"gopkg.in/yaml.v3"
)

type Type struct {
	Kind      TypeKind `json:"kind"`
	KeyType   *Type    `json:"key_type"`  // map
	ValueType *Type    `json:"elem_type"` // collection
}

type TypeKind int

const (
	IntType TypeKind = iota
	StringType
	ArrayType
	ObjectType
)

func (t Type) String() string {
	switch t.Kind {
	case IntType:
		return "int"
	case StringType:
		return "string"
	case ArrayType:
		return "array<" + t.ValueType.String() + ">"
	case ObjectType:
		return "object<" + t.KeyType.String() + "," + t.ValueType.String() + ">"
	}
	panic("unreachable")
}

type FunctionDeclaration struct {
	Name       string  `yaml:"name"`
	Raw        bool    `yaml:"bool"`
	Parameters []*Type `yaml:"parameters"`
	Returns    []*Type `yaml:"returns"`
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
