package main

import (
	"github.com/spf13/viper"
)

// Config used to load configuration
type Config struct {
	storageDirectory string
	authEnabled      bool
	username         string
	password         string
}

func (c *Config) loadConfig(envfile string) {
	viper.SetConfigFile(".env.dist")
	viper.SetConfigType("env")
	viper.AutomaticEnv()

	viper.SetDefault("tf_storage_dir", "./store")
	viper.SetDefault("tf_auth_enabled", false)
	viper.SetDefault("tf_username", "")
	viper.SetDefault("tf_password", "")

	if err := viper.ReadInConfig(); err != nil {
		logger.Infof("Error while reading config file %s", err)
	}

	viper.SetConfigFile(envfile)
	if err := viper.MergeInConfig(); err != nil {
		logger.Warnf("Error while reading config file %s", err)
	}

	c.storageDirectory = viper.GetString("tf_storage_dir")
	c.authEnabled = viper.GetBool("tf_auth_enabled")
	c.username = viper.GetString("tf_username")
	c.password = viper.GetString("tf_password")
}

func (c *Config) getAuthMap() map[string]string {
	authData := make(map[string]string)

	authData[c.username] = c.password

	return authData
}
