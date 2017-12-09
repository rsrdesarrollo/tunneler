package main

import (
	log "github.com/alecthomas/log4go"
	flags "github.com/jessevdk/go-flags"

	"errors"
	"fmt"
	"regexp"

	"github.com/gorilla/websocket"
	"github.com/pkg/profile"
	"github.com/rsrdesarrollo/tunneler/common"
	"github.com/rsrdesarrollo/tunneler/messages"
	"os"
	"os/signal"
)

func main() {
	err := Run()

	if err != nil {
		fmt.Println("ERROR: ", err)
	}
}

var logger log.Logger

type Tunnel struct {
	Protocol       string
	BindService    string
	ConnectService string
}

var options struct {
	Verbose      []bool `short:"v" long:"verbose" description:"Show verbose debug information"`
	Server       string `short:"s" long:"server" description:"tunneler server address" default:"ws://localhost:8080/ws"`
	RemoteTunnel string `short:"R" description:"remote tunnel address"`
	LocalTunnel  string `short:"L" description:"local tunnel address"`
	Profile      bool   `long:"profile" description:"profile application"`
	Protocol     string `short:"p" description:"tunnel protocol (tcp/udp)" default:"tcp" choice:"tcp" choice:"udp"`
}

func parseTunnelString(protocol string, tunnel string) (*Tunnel, error) {
	tunnelRegex := regexp.MustCompile(`^(?P<BindService>(?:[^:]+:)?[^:]+):(?P<ConnectService>[^:]+:[^:]+)$`)

	ok := tunnelRegex.MatchString(tunnel)

	if ok {
		match := tunnelRegex.FindStringSubmatch(tunnel)

		return &Tunnel{
			Protocol:       protocol,
			BindService:    match[1],
			ConnectService: match[2],
		}, nil
	} else {
		return nil, errors.New("invalid tunnel format")
	}
}

func Run() error {

	flags.Parse(&options)

	if options.Profile {
		defer profile.Start().Stop()
	}

	if len(options.Verbose) > 3 {
		logger = log.NewDefaultLogger(log.DEBUG)
	} else if len(options.Verbose) == 3 {
		logger = log.NewDefaultLogger(log.TRACE)
	} else if len(options.Verbose) == 2 {
		logger = log.NewDefaultLogger(log.INFO)
	} else if len(options.Verbose) == 1 {
		logger = log.NewDefaultLogger(log.WARNING)
	} else {
		logger = log.NewDefaultLogger(log.ERROR)
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

	ws, _, err := dialer.Dial(options.Server, nil)

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

func createRemoteTunnel(wsocket *websocket.Conn, tunnel *Tunnel) error {
	logger.Debug("createRemoteTunnel")

	exitPoint, err := common.NewExitPoint(wsocket, tunnel.Protocol, tunnel.ConnectService, logger)
	if err != nil {
		return err
	}

	exitPoint.WebsocketWritterChannel <- messages.CreateRemoteTunnelMessage(tunnel.Protocol, tunnel.BindService)
	response := <-exitPoint.WebsocketReaderChannel

	if response.Type == messages.MessageType.Error {
		return errors.New(response.Description)
	} else if response.Type == messages.MessageType.RemoteTunnelReady {
		logger.Info("Remote tunnel binded on %s://%s", response.Protocol, response.Service)
	} else {
		return errors.New("protocol mistmach")
	}

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	go func() {
		for _ = range c {
			exitPoint.CloseChannel()
		}
	}()

	<-exitPoint.Done

	return nil
}

func createLocalTunnel(wsocket *websocket.Conn, tunnel *Tunnel) error {
	logger.Debug("createLocalTunnel")

	entryPoint, err := common.NewEntryPoint(wsocket, tunnel.Protocol, tunnel.BindService, logger)
	if err != nil {
		return err
	}

	entryPoint.WebsocketWritterChannel <- messages.CreateLocalTunnelMessage(tunnel.Protocol, tunnel.ConnectService)
	response := <-entryPoint.WebsocketReaderChannel

	if response.Type == messages.MessageType.Error {
		return errors.New(response.Description)
	} else if response.Type == messages.MessageType.LocalTunnelReady {
		logger.Info("Local tunnel binded on %s://%s", response.Protocol, response.Service)
	} else {
		return errors.New("protocol mistmach")
	}

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	go func() {
		for _ = range c {
			entryPoint.CloseChannel()
		}
	}()

	<-entryPoint.Done

	return nil
}
