package main

import (
	"github.com/rsrdesarrollo/tunneler/messages"
	"os"
	"os/signal"
	"github.com/gorilla/websocket"
	"github.com/rsrdesarrollo/tunneler/common"
	"errors"
	"regexp"
)

type Tunnel struct {
	Protocol       string
	BindService    string
	ConnectService string
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
