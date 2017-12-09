package common

import (
	"encoding/json"
	"github.com/alecthomas/log4go"
	"github.com/gorilla/websocket"
	"github.com/rsrdesarrollo/tunneler/messages"
	"net"
	"sync"
	"time"
)

type ExitPoint struct {
	Websocket               *websocket.Conn
	Clients                 map[string]*Client
	Service                 string
	Protocol                string
	WebsocketWritterChannel chan *messages.Message
	WebsocketReaderChannel  chan *messages.Message
	Done                    chan bool

	mutex  sync.Mutex
	isOpen bool
	log    log4go.Logger
}

func NewExitPoint(wsocket *websocket.Conn, protocol string, service string, log log4go.Logger) (*ExitPoint, error) {

	obj := &ExitPoint{
		Websocket:               wsocket,
		Clients:                 make(map[string]*Client),
		Service:                 service,
		Protocol:                protocol,
		WebsocketWritterChannel: make(chan *messages.Message, 10),
		WebsocketReaderChannel:  make(chan *messages.Message, 10),

		Done: make(chan bool, 1),

		mutex:  sync.Mutex{},
		isOpen: true,
		log:    log,
	}

	go obj.WebsocketWriter()
	go obj.WebsocketReader()

	return obj, nil
}

func (self *ExitPoint) ReceiveDataFromWebsocket(clientId string, data []byte) {
	self.log.Debug("ReceiveDataFromWebsocket")

	self.mutex.Lock()
	client := self.Clients[clientId]
	if client == nil {
		self.log.Trace("Client %s not connected. connecting to %s", clientId, self.Service)

		connection, err := net.DialTimeout(self.Protocol, self.Service, 60*time.Second)

		if err != nil {
			// TODO: Handle error on connection failed.
			self.log.Error(err)
			self.mutex.Unlock()
			return
		}

		client = NewClient(
			clientId,
			connection,
			self.log,
		)
		self.Clients[clientId] = client
		go client.ClientHandler(self)
	}

	writeLen, err := client.connection.Write(data)

	self.mutex.Unlock()

	self.log.Trace("Client %s, write %d bytes", clientId, writeLen)

	if err != nil {
		//TODO handle error but not EOF
		self.log.Error(err)
		self.CloseClient(clientId)
	}
}

func (self *ExitPoint) ReceiveDataFromClientSocket(client *Client, data []byte) {
	self.log.Debug("ReceiveDataFromClientSocket")
	self.WebsocketWritterChannel <- messages.DataMessage(
		client.id,
		data,
	)
}

func (self *ExitPoint) WebsocketWriter() {
	var msg *messages.Message

	self.log.Debug("WebsocketWritter")

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

func (self *ExitPoint) WebsocketReader() {
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

		// TODO Handle Error messages

		if msg.Type == messages.MessageType.Data {
			if len(msg.Data) == 0 {
				self.log.Trace("Client %s, received EOF from websocket", msg.ClientId)
				self.CloseClient(msg.ClientId)
				continue
			}

			self.ReceiveDataFromWebsocket(msg.ClientId, msg.Data)
		} else {
			self.WebsocketReaderChannel <- msg
		}

	}
}

func (self *ExitPoint) CloseClient(clientId string) {
	self.log.Debug("CloseClient")

	client := self.Clients[clientId]
	self.Clients[clientId] = nil

	if client != nil {
		self.log.Trace("Closing client %s", clientId)
		client.connection.Close()
	}
}

func (self *ExitPoint) CloseChannel() {
	self.log.Debug("CloseChannel")
	self.isOpen = false

	for clientId, _ := range self.Clients {
		self.CloseClient(clientId)
	}

	self.Done <- true
}

func (self *ExitPoint) TerminateChannel(error error) {
	self.log.Critical(error)
	self.CloseChannel()
	self.Websocket.Close()
}

func (self *ExitPoint) IsOpen() bool {
	return self.isOpen
}
