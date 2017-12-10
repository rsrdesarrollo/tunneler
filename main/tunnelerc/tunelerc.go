package main

import (
	log "github.com/alecthomas/log4go"
	flags "github.com/jessevdk/go-flags"

	"errors"

	"github.com/gorilla/websocket"
	"github.com/pkg/profile"
	"github.com/rsrdesarrollo/tunneler/aux"
	"github.com/spf13/viper"
	"net/http"
)

var logger log.Logger

var options struct {
	RemoteTunnel string `short:"R" description:"remote tunnel address"`
	LocalTunnel  string `short:"L" description:"local tunnel address"`
	Profile      bool   `long:"profile" description:"profile application"`
	Protocol     string `short:"p" description:"tunnel protocol (tcp/udp)" default:"tcp" choice:"tcp" choice:"udp"`
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

	logger = log.NewDefaultLogger(aux.LogLevel(viper.GetString("LogLevel")))

	if !viper.IsSet("Token") {
		return errors.New("need to specify Token in configuration")
	}

	return nil
}

func run() error {
	if options.Profile {
		defer profile.Start().Stop()
	}

	if options.RemoteTunnel != "" && options.LocalTunnel != "" {
		return errors.New("unable to create local and remote tunnel at the same time")
	}

	//urlProxy, _ := url.Parse("http://127.0.0.1:8080")
	dialer := websocket.Dialer{
		//Proxy:           http.ProxyURL(urlProxy),
		WriteBufferSize: 40960,
		ReadBufferSize:  40960,
	}

	ws, _, err := dialer.Dial(viper.GetString("Server"), http.Header{
		"Authorization": {"Bearer " + viper.GetString("Token")},
	})

	if err != nil {
		return err
	}

	tunnelStr := options.RemoteTunnel
	if tunnelStr == "" {
		tunnelStr = options.LocalTunnel
	}

	if options.RemoteTunnel != "" {
		tunnel, err := parseTunnelString(options.Protocol, tunnelStr)
		if err != nil {
			return err
		}

		err = createRemoteTunnel(ws, tunnel)
		if err != nil {
			return err
		}
	} else if options.LocalTunnel != "" {
		tunnel, err := parseTunnelString(options.Protocol, tunnelStr)
		if err != nil {
			return err
		}

		err = createLocalTunnel(ws, tunnel)
		if err != nil {
			return err
		}
	} else {
		return errors.New("need at least one type of tunnel")
	}

	return err
}