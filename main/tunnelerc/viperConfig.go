package main

import (
	"github.com/spf13/viper"
	"fmt"
)

func configureViper() {
	viper.SetConfigType("yaml")

	viper.SetDefault("LogLevel", "INFO")

	viper.SetDefault("ReadBufferSize", 40960)
	viper.SetDefault("WriteBufferSize", 40960)
	viper.SetDefault("ClientSocketBuffer", 40960)

	viper.SetDefault("Server", "ws://127.0.0.1:9000")
	viper.BindEnv("HttpProxy", "http_proxy")
	viper.RegisterAlias("HttpProxy", "Proxy")

	viper.SetConfigName("config")
	viper.AddConfigPath("/etc/tunnelerc/")
	viper.AddConfigPath("$HOME/.tunnelerc")
	viper.AddConfigPath(".")
	err := viper.ReadInConfig()

	if err != nil {
		panic(fmt.Errorf("Fatal error config file: %s \n", err))
	}
}
