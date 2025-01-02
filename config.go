package main

import (
	"encoding/json"
	"fmt"
	"os"
)

type DirectoryConfig struct {
	Name   string `json:"name`
	Type   string `json:"type` // "s3" or "local"
	Bucket string `json:"bucket,omitempty`
	Prefix string `json:"prefix,omitempty`
	Path   string `json:"bucket,omitempty`
}

type Config struct {
	Directories []DirectoryConfig `json:"directories"`
}

func CargarConfiguracion(path string) (Config, error) {
	var config Config

	data, err := os.ReadFile(path)
	if err != nil {
		return config, fmt.Errorf("error leyendo archivo de configuración: %w", err)
	}

	if err := json.Unmarshal(data, &config); err != nil {
		return config, fmt.Errorf("error parsing JSON: %w", err)
	}

	return config, nil
}
