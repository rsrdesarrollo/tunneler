package common

import (
	"encoding/json"
	"github.com/alecthomas/log4go"
	"github.com/gorilla/websocket"
	"github.com/rsrdesarrollo/tunneler/messages"
	"net"
	"strconv"
	"sync"
)

type EntryPoint struct {
	Websocket               *websocket.Conn
	Clients                 map[string]*Client
	Service                 string
	Protocol                string
	WebsocketWritterChannel chan *messages.Message
	WebsocketReaderChannel  chan *messages.Message
	Listener                net.Listener
	Done                    chan bool

	mutex  sync.Mutex
	isOpen bool
	log    log4go.Logger
}

func NewEntryPoint(wsocket *websocket.Conn, protocol string, service string, log log4go.Logger) (*EntryPoint, error) {

	// TODO: implement UDP (might change a lot of things)
	listener, err := net.Listen(protocol, service)

	if err != nil {
		return nil, err
	}

	obj := &EntryPoint{
		Websocket:               wsocket,
		Clients:                 make(map[string]*Client),
		Service:                 service,
		Protocol:                protocol,
		WebsocketReaderChannel:  make(chan *messages.Message, 10),
		WebsocketWritterChannel: make(chan *messages.Message, 10),
		Listener:                listener,
		Done:                    make(chan bool, 1),

		mutex:  sync.Mutex{},
		isOpen: true,
		log:    log,
	}

	obj.log.Info("Entry point binded on %s", service)

	//Connection handler loop
	go obj.ConnectionHandler()
	// Reader Loop
	go obj.WebsocketReader()
	// Writer Loop
	go obj.WebsocketWriter()

	return obj, nil
}

func (self *EntryPoint) ReceiveDataFromWebsocket(client *Client, data []byte) {
	self.log.Debug("ReceiveDataFromWebsocket")

	if client == nil {
		self.log.Warn("Receiving data from unexsiten or closed client.")
		return
	}

	self.log.Debug("client: %s, data: %s", client.id, data)

	writeLen, err := client.connection.Write(data)

	self.log.Trace("Client %s, write %d bytes", client.id, writeLen)

	if err != nil {
		//TODO handle error but not EOF
		self.log.Error(err)
		self.CloseClient(client.id)
	}
}

func (self *EntryPoint) ReceiveDataFromClientSocket(client *Client, data []byte) {
	self.WebsocketWritterChannel <- messages.DataMessage(
		client.id,
		data,
	)
}

func (self *EntryPoint) WebsocketWriter() {
	self.log.Debug("WebsocketWriter")
	var msg *messages.Message
	for self.isOpen {
		msg = <-self.WebsocketWritterChannel

		msgJson, _ := json.Marshal(msg)
		self.log.Trace("Writting message to websocket %s", msgJson)

		err := self.Websocket.WriteJSON(msg)
		if err != nil {
			self.TerminateChannel(err)
			break // TODO: call entrypoint terminate
		}
	}
}

func (self *EntryPoint) WebsocketReader() {
	self.log.Debug("WebsocketReader")

	for self.isOpen {
		msg := messages.New()

		err := self.Websocket.ReadJSON(msg)

		if err != nil {
			self.TerminateChannel(err)
			break // TODO: call entrypoint terminate
		}

		msgJson, _ := json.Marshal(msg)
		self.log.Trace("Readed message from websocket %s", msgJson)

		if msg.Type == messages.MessageType.Data {
			client := self.Clients[msg.ClientId]
			if len(msg.Data) == 0 {
				if client.readyToClose {
					self.log.Trace("Client %s, received EOF from websocket", msg.ClientId)
					self.CloseClient(msg.ClientId)
					continue
				} else {
					client.readyToClose = true
					continue
				}
			}
			self.ReceiveDataFromWebsocket(client, msg.Data)
		} else {
			self.WebsocketReaderChannel <- msg
		}
	}
}

func (self *EntryPoint) ConnectionHandler() {
	self.log.Debug("ConnectionHandler")
	var clientIdGenerator = 1 //TODO: Make something more robust?
	for self.isOpen {
		connection, err := self.Listener.Accept()

		if err != nil {
			self.TerminateChannel(err)
			break // TODO: call entrypoint terminate
		}

		self.mutex.Lock()

		clientId := strconv.Itoa(clientIdGenerator)
		clientIdGenerator++

		self.log.Info("New client [%s] from %s", clientId, connection.RemoteAddr().String())

		self.Clients[clientId] = NewClient(
			clientId,
			connection,
			self.log,
		)

		self.mutex.Unlock()

		// Handle client read data
		go self.Clients[clientId].ClientHandler(self)
	}
}

func (self *EntryPoint) CloseClient(clientId string) {

	client := self.Clients[clientId]
	self.Clients[clientId] = nil

	if client != nil {
		client.connection.Close()
	}
}

func (self *EntryPoint) CloseChannel() {
	self.isOpen = false

	for clientId, _ := range self.Clients {
		self.CloseClient(clientId)
	}

	self.Done <- true
}

func (self *EntryPoint) TerminateChannel(error error) {
	self.log.Critical(error)
	self.Listener.Close()
	self.CloseChannel()
	self.Websocket.Close()
}

func (self *EntryPoint) IsOpen() bool {
	return self.isOpen
}
