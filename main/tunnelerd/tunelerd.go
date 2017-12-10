package main

import (
	"net/http"

	"github.com/pkg/profile"

	log "github.com/alecthomas/log4go"
	flags "github.com/jessevdk/go-flags"

	"github.com/rsrdesarrollo/tunneler/aux"

	"github.com/spf13/viper"
	"time"
	"errors"
	"fmt"
	"os"
)

var logger log.Logger
var secret_key []byte
var version = "undefined"

var options struct {
	PrintVersion  bool   `long:"version" description:"print version and exit"`
	Profile        bool   `long:"profile" description:"profile application"`
	GenerateToken  bool   `long:"generate-token" description:"run token generation for a user and exit"`
	User           string `short:"u" long:"user" description:"username for the token to be generated"`
	ExpirationTime int64  `short:"e" long:"expiration" description:"expiration time of the token in days" default:"360"`
}

func main() {

	configureViper()

	err := initialize()

	defer logger.Close()

	if err != nil {
		logger.Critical(err)
		return
	}

	err = run()

	if err != nil {
		logger.Critical(err)
		return
	}

}

func initialize() error {
	flags.Parse(&options)

	if options.PrintVersion{
		fmt.Printf("Version: %s\n", version)
		os.Exit(0)
	}

	logger = log.NewDefaultLogger(aux.LogLevel(viper.GetString("LogLevel")))

	if !viper.IsSet("SecretKey") {
		return errors.New("need to specify SecretKey in configuration")
	}

	secret_key = []byte(viper.GetString("SecretKey"))

	return nil
}

func run() error {

	if options.Profile {
		defer profile.Start().Stop()
	}

	if options.GenerateToken {
		token, err := generateToken(options.User, time.Duration(options.ExpirationTime))

		if err != nil {
			return err
		}

		logger.Info("User %s, token generated '%s'", options.User, token)

		return nil
	}

	http.HandleFunc("/ws", validToken(serveWebsocket))

	bindAddress := viper.GetString("BindAddress")
	logger.Info("Listen on address %s.", bindAddress)

	return http.ListenAndServe(bindAddress, nil)
}
