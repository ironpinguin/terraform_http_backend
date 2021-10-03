package main

import (
	"log"

	"github.com/spf13/viper"
)

type Config struct {
	storageDirectory string
	authEnabled      bool
	username         string
	password         string
}

func (c *Config) loadConfig(envfile string) {
	viper.SetConfigFile(".env")
	viper.SetEnvPrefix("tfhttp")
	viper.AutomaticEnv()

	viper.SetDefault("storage_dir", "./store")
	viper.SetDefault("auth_enabled", false)
	viper.SetDefault("username", "")
	viper.SetDefault("password", "")

	err := viper.ReadInConfig()
	if err != nil {
		log.Fatalf("Error while reading config file %s", err)
	}
	c.storageDirectory = viper.GetString("storage_dir")
	c.authEnabled = viper.GetBool("auth_enabled")
	c.username = viper.GetString("username")
	c.password = viper.GetString("password")
}

func (c *Config) getAuthMap() map[string]string {
	authData := make(map[string]string)

	authData[c.username] = c.password

	return authData
}
