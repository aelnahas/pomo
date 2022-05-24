package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/BurntSushi/toml"
)

const DefaultPath = "config.toml"

type Config struct {
	Database Database    `toml:"database"`
	Timers   TimerConfig `toml:"timers"`
}

type Database struct {
	Task    string `toml:"task"`
	Session string `toml:"session"`
}

func ExpandPath(path string) (string, error) {
	if strings.HasPrefix(path, "~/") {
		dirname, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		return filepath.Join(dirname, path[2:]), nil
	}
	return path, nil
}

type TimerConfig struct {
	Focus    string `toml:"focus"`
	Short    string `toml:"short"`
	Long     string `toml:"long"`
	Interval int    `toml:"interval"`
}

func (tc *TimerConfig) FocusDuration() time.Duration {
	d, err := time.ParseDuration(tc.Focus)
	if err != nil {
		panic(err)
	}
	return d
}

func (tc *TimerConfig) ShortBreakDuration() time.Duration {
	d, err := time.ParseDuration(tc.Short)
	if err != nil {
		panic(err)
	}
	return d
}

func (tc *TimerConfig) LongBreakDuration() time.Duration {
	d, err := time.ParseDuration(tc.Long)
	if err != nil {
		panic(err)
	}
	return d
}

type duration struct {
	time.Duration
}

func (d *duration) UnmarshalText(text []byte) (err error) {
	d.Duration, err = time.ParseDuration(string(text))
	return err
}

func (d *duration) MarshalText() ([]byte, error) {
	text := d.Duration.String()
	return []byte(text), nil
}

func Parse(path string) (*Config, error) {
	if _, err := os.Stat(path); err != nil {
		path = "config.toml"
	}

	var config Config
	_, err := toml.DecodeFile(path, &config)
	if err != nil {
		return nil, err
	}

	return &config, nil
}

func Update(key, value string, config *Config) error {
	if strings.HasPrefix(key, "database.") {
		switch {
		case strings.HasPrefix(key, "task"):
			config.Database.Task = value
		case strings.HasPrefix(key, "session"):
			config.Database.Session = value
		default:
			return fmt.Errorf("unknown database %s", key)
		}
	} else if strings.HasPrefix(key, "timers") {
		switch {
		case strings.HasSuffix(key, "focus"):
			d, err := time.ParseDuration(value)
			if err != nil {
				return err
			}
			fmt.Println(d, value)
			config.Timers.Focus = d.String()
		case strings.HasSuffix(key, "short"):
			d, err := time.ParseDuration(value)
			if err != nil {
				return err
			}
			config.Timers.Short = d.String()
		case strings.HasSuffix(key, "long"):
			d, err := time.ParseDuration(value)
			if err != nil {
				return err
			}
			config.Timers.Long = d.String()
		case strings.HasSuffix(key, "intervals"):
			intervals, err := strconv.Atoi(value)
			if err != nil {
				return err
			}
			config.Timers.Interval = intervals
		default:
			return fmt.Errorf("unknown key %s", key)
		}
	} else {
		return fmt.Errorf("unknown key %s", key)
	}

	f, err := os.Create(DefaultPath)
	if err != nil {
		return err
	}
	defer f.Close()

	return toml.NewEncoder(f).Encode(config)
}
