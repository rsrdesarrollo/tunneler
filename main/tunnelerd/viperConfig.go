package main

import (
	"github.com/spf13/viper"
	"fmt"
)

func configureViper() {
	viper.SetConfigType("yaml")

	viper.SetDefault("LogLevel", "INFO")
	viper.SetDefault("BindAddress", "127.0.0.1:9000")

	viper.SetConfigName("config")
	viper.AddConfigPath("/etc/tunnelerd/")
	viper.AddConfigPath("$HOME/.tunnelerd")
	viper.AddConfigPath(".")
	err := viper.ReadInConfig()

	if err != nil {
		panic(fmt.Errorf("Fatal error config file: %s \n", err))
	}
}
