package common

import (
	"github.com/alecthomas/log4go"
	"io"
	"net"
)

type Client struct {
	id           string
	connection   net.Conn
	log          log4go.Logger
	readyToClose bool
}

func NewClient(id string, connetcion net.Conn, log log4go.Logger) *Client {
	return &Client{
		id:           id,
		connection:   connetcion,
		log:          log,
		readyToClose: false,
	}
}

func (self *Client) ClientHandler(point TunnelPoint) {
	self.log.Debug("ClientHandler")

	defer self.connection.Close()

	buffer := make([]byte, 40960)

	for point.IsOpen() {
		readLen, err := self.connection.Read(buffer)

		self.log.Trace("Client %s, read %d bytes '%s'", self.id, readLen, buffer[:readLen])

		// TODO: Think something less memory heap cookie monster (e.g. circular buffer...)
		data := make([]byte, readLen)
		copy(data, buffer[:readLen])

		point.ReceiveDataFromClientSocket(self, data)

		if err == io.EOF {
			if self.readyToClose {
				point.CloseClient(self.id)
			} else {
				self.readyToClose = true
			}
			break
		}

		if err != nil {
			self.log.Error(err)
			point.CloseClient(self.id)
			break
		}
	}
}
