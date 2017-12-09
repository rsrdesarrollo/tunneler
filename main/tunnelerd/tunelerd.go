package main

import (
	"fmt"
	"github.com/rsrdesarrollo/tunneler/messages"
	"net/http"

	auth "github.com/abbot/go-http-auth"
	flags "github.com/jessevdk/go-flags"

	"errors"
	log "github.com/alecthomas/log4go"
	"github.com/gorilla/websocket"
	"github.com/pkg/profile"
	"github.com/rsrdesarrollo/tunneler/common"
)

func main() {
	err := Run()

	if err != nil {
		fmt.Println("ERROR: ", err)
	}
}

var logger log.Logger

var options struct {
	Verbose      []bool `short:"v" long:"verbose" description:"Show verbose debug information"`
	Address      string `short:"a" long:"address" description:"http service address" default:"127.0.0.1:8080"`
	HtpasswdFile string `short:"p" long:"htpasswd" description:"htpassword file" value-name:"FILE"`
	Profile      bool   `long:"profile" description:"profile application"`
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

	if options.HtpasswdFile != "" {
		authenticator := auth.NewBasicAuthenticator("tunneler", auth.HtpasswdFileProvider(options.HtpasswdFile))
		http.HandleFunc("/ws", authenticator.Wrap(serveWebsocketAuthenticated))
	} else {
		logger.Warn("Service runing without authentication.")
		http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
			authReq := &auth.AuthenticatedRequest{Request: *r, Username: "anonymous"}
			serveWebsocketAuthenticated(w, authReq)
		})
	}

	logger.Info("Listen on address %s.", options.Address)

	return http.ListenAndServe(options.Address, nil)
}

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
	WriteBufferSize: 40960,
	ReadBufferSize:  40960,
}

func serveWebsocketAuthenticated(w http.ResponseWriter, request *auth.AuthenticatedRequest) {
	logger.Debug("serveWebsocket")
	ws, err := upgrader.Upgrade(w, &request.Request, nil)

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

	logger.Trace("(%s) Readed message from websocket %#v", request.Username, msg)

	if msg.Type == messages.MessageType.CreateLocalTunnel {
		logger.Debug("(%s) Client ask to create a Local Tunnel", request.Username)
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
		log.Debug("Local tunnel Done.")

	} else if msg.Type == messages.MessageType.CreateRemoteTunnel {
		logger.Debug("(%s) Client ask to create a Remote Tunnel", request.Username)
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
		log.Debug("Remote tunnel Done.")

	} else {
		ws.WriteJSON(messages.ErrorMessage(errors.New("protocol mismatch")))
	}
}
