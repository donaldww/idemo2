// Copyright 2019 by Donald Wilson. All rights reserved.
// Use of this source code is governed by an MIT
// license that can be found in the LICENSE file.

package conf

import (
	"fmt"
	"time"
	
	"github.com/spf13/viper"
)

// Config is a context handle to a toml config file.
type Config int

// NewConfig returns a config file context.
func NewConfig(configFile string, path ...string) Config {
	// configFile=<a_file_name> without the filename extension.
	viper.SetConfigName(configFile)
	
	// AddConfigPath may be called multiple times to add directories.
	viper.AddConfigPath(".")
	viper.AddConfigPath("./testdata")
	if path != nil {
		viper.AddConfigPath(path[0])
	}
	
	err := viper.ReadInConfig()
	if err != nil {
		panic(fmt.Errorf("Fatal error reading config file: %s\n", err))
	}
	
	//TODO: implement WatchConfig reload
	// viper.WatchConfig()
	// viper.OnConfigChange(reload)
	return *new(Config)
}

// GetInt returns an int from the infinigongroup Config file.
func (c Config) GetInt(key string) int {
	return viper.GetInt(key)
}

// GetString returns an int from the infinigongroup Config file.
func (c Config) GetString(key string) string {
	return viper.GetString(key)
}

// GetBool returns an int from the infinigongroup Config file.
func (c Config) GetBool(key string) bool {
	return viper.GetBool(key)
}

// GetFloat64 returns an int from the infinigongroup Config file.
func (c Config) GetFloat64(key string) float64 {
	return viper.GetFloat64(key)
}

// GetMilliseconds returns a Duration in milliseconds.
func (c Config) GetMilliseconds(key string) time.Duration {
	return time.Duration(viper.GetInt(key)) * time.Millisecond
}

// GetSeconds returns a Duration in seconds.
func (c Config) GetSeconds(key string) time.Duration {
	return time.Duration(viper.GetInt(key)) * time.Second
}
