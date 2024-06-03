package config

import (
	"fmt"
	"log"
	"os"
	"sync"
	"time"

	"github.com/spf13/viper"
)

var (
	once sync.Once
	conf *Config
)

type Config struct {
	home string
}

// NewConfig initializes and returns a new Config instance.
func NewConfig(filename string) *Config {
	once.Do(func() {
		home := getHome()
		viper.SetConfigName(filename)
		viper.AddConfigPath(home)
		err := viper.ReadInConfig()
		if err != nil {
			panic(fmt.Errorf("fatal error config file: %s", err))
		}
		conf = &Config{home: home}
	})
	return conf
}

// GetInt returns an int from the config file.
func (c *Config) GetInt(key string) int {
	return viper.GetInt(key)
}

// GetString returns a string from the config file.
func (c *Config) GetString(key string) string {
	return viper.GetString(key)
}

// GetBool returns a bool from the config file.
func (c *Config) GetBool(key string) bool {
	return viper.GetBool(key)
}

// GetFloat64 returns a float64 from the config file.
func (c *Config) GetFloat64(key string) float64 {
	return viper.GetFloat64(key)
}

// GetMilliseconds returns a Duration in milliseconds.
func (c *Config) GetMilliseconds(key string) time.Duration {
	return time.Duration(viper.GetInt(key)) * time.Millisecond
}

// GetSeconds returns a Duration in seconds.
func (c *Config) GetSeconds(key string) time.Duration {
	return time.Duration(viper.GetInt(key)) * time.Second
}

// Home returns the home directory.
func (c *Config) Home() string {
	return c.home
}

// Bin returns config bin directory.
func (c *Config) Bin() string {
	return c.home + "/bin"
}

func getHome() string {
	home, err := os.UserHomeDir()
	if err != nil {
		log.Fatal(err)
	}
	return home + "/.config/enclave"
}
