package config

import (
	"costrict-keeper/internal/env"
	"costrict-keeper/internal/logger"
	"costrict-keeper/internal/models"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
)

func loadLocalSpec() (*models.SystemSpecification, error) {
	fname := filepath.Join(env.CostrictDir, "share", "system-spec.json")

	bytes, err := os.ReadFile(fname)
	if err != nil {
		return nil, fmt.Errorf("load 'system-spec.json' failed: %v", err)
	}
	var spec models.SystemSpecification
	if err := json.Unmarshal(bytes, &spec); err != nil {
		return nil, fmt.Errorf("unmarshal 'system-spec.json' failed: %v", err)
	}
	return &spec, nil
}

var system *models.SystemSpecification

func LoadSpec() error {
	if system != nil {
		return nil
	}
	var err error
	system, err = loadLocalSpec()
	if err != nil {
		logger.Errorf("Load failed: %v", err)
		return err
	}
	return nil
}

func Spec() *models.SystemSpecification {
	if system == nil {
		log.Fatalln("Must run config.LoadSpec first")
		return nil
	}
	return system
}
