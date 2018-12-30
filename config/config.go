package config

import (
	"encoding/json"
	"os"
)

// Cfg represents a configuration file
type Cfg struct {
	Host string
}

// Read reads a config file
func (c *Cfg) Read(file string) error {
	f, err := os.Open(file)
	if err != nil {
		return err
	}

	decoder := json.NewDecoder(f)

	err = decoder.Decode(&c)
	if err != nil {
		return err
	}

	return nil
}
