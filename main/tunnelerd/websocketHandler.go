package main

import (
	"net/http"
	"errors"

	"github.com/rsrdesarrollo/tunneler/messages"
	"github.com/gorilla/websocket"
	"github.com/rsrdesarrollo/tunneler/common"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
	WriteBufferSize: 40960,
	ReadBufferSize:  40960,
}

func serveWebsocket(w http.ResponseWriter, request *http.Request) {

	logger.Debug("serveWebsocket")
	ws, err := upgrader.Upgrade(w, request, nil)

	if err != nil {
		logger.Error(err)
		return
	}

	defer ws.Close()

	// Handle tunnel handshake
	msg := messages.New()
	err = ws.ReadJSON(msg)

	if err != nil {
		logger.Error(err)
		return
	}

	logger.Trace("(%s) Readed message from websocket %#v", "", msg)

	if msg.Type == messages.MessageType.CreateLocalTunnel {
		logger.Debug("(%s) Client ask to create a Local Tunnel", "")
		exitPoint, err := common.NewExitPoint(
			ws,
			msg.Protocol,
			msg.Service,
			logger,
		)

		if err != nil {
			//TODO: handle create exit point error
			logger.Error(err)
		}

		exitPoint.WebsocketWritterChannel <- messages.LocalTunnelReadyMessage(msg.Protocol, msg.Service)

		<-exitPoint.Done
		logger.Debug("Local tunnel Done.")

	} else if msg.Type == messages.MessageType.CreateRemoteTunnel {
		logger.Debug("(%s) Client ask to create a Remote Tunnel", "")
		entryPoint, err := common.NewEntryPoint(
			ws,
			msg.Protocol,
			msg.Service,
			logger,
		)

		if err != nil {
			// TODO: handle create entry point error
			logger.Error(err)
		}

		entryPoint.WebsocketWritterChannel <- messages.RemoteTunnelReadyMessage(msg.Protocol, msg.Service)

		<-entryPoint.Done
		logger.Debug("Remote tunnel Done.")

	} else {
		ws.WriteJSON(messages.ErrorMessage(errors.New("protocol mismatch")))
	}
}
