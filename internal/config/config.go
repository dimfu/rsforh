package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"reflect"
)

type Key = string

const (
	RBRInstallationPath Key = "RBRInstallationPath"
	DiscordToken        Key = "DiscordToken"
)

type Config struct {
	RBRInstallationPath Key `json:"rbr_installation_path"`
	ConfigPath          Key `json:"configPath"`
	DiscordToken        Key `json:"discord_token"`
}

func (c *Config) Get(name Key) (string, error) {
	v := reflect.ValueOf(c).Elem()
	f := v.FieldByName(name)

	if f.Len() == 0 {
		return "", fmt.Errorf("key %q value is empty", name)
	}

	return f.String(), nil
}

func (c *Config) Set(name string, value any) error {
	v := reflect.ValueOf(c).Elem()
	f := v.FieldByName(name)

	if !f.IsValid() {
		return fmt.Errorf("no such field: %s", name)
	}
	if !f.CanSet() {
		return fmt.Errorf("cannot set field: %s", name)
	}

	val := reflect.ValueOf(value)

	if !val.Type().AssignableTo(f.Type()) {
		if val.Type().ConvertibleTo(f.Type()) {
			val = val.Convert(f.Type())
		} else {
			return fmt.Errorf("cannot assign value of type %s to field %s of type %s",
				val.Type(), name, f.Type())
		}
	}

	f.Set(val)

	cfgPath := path.Join(c.ConfigPath, "config.json")
	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return err
	}
	if err := os.WriteFile(cfgPath, data, 0600); err != nil {
		return err
	}
	return nil
}

func initConfig() (*Config, error) {
	userCfgPath, err := os.UserConfigDir()
	if err != nil {
		return nil, err
	}
	configDir := filepath.Join(userCfgPath, "rsforh")

	if err := os.MkdirAll(configDir, 0755); err != nil {
		return nil, err
	}

	configFile := filepath.Join(configDir, "config.json")

	var cfg Config

	if _, err := os.Stat(configFile); os.IsNotExist(err) {
		cfg = Config{
			RBRInstallationPath: "",
			ConfigPath:          configDir,
		}

		if err := writeConfig(configFile, &cfg); err != nil {
			return nil, err
		}
	} else {
		f, err := os.Open(configFile)
		if err != nil {
			return nil, err
		}
		defer f.Close()

		if err := json.NewDecoder(f).Decode(&cfg); err != nil {
			return nil, err
		}
	}

	return &cfg, nil
}

func writeConfig(path string, cfg *Config) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()

	encoder := json.NewEncoder(f)
	encoder.SetIndent("", "  ")
	return encoder.Encode(cfg)
}

func Setup() (*Config, error) {
	var err error
	cfg, err := initConfig()
	if err != nil {
		return nil, err
	}
	return cfg, nil
}
