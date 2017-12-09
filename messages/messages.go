package messages

import (
	"fmt"
)

func New() *Message {
	return &Message{}
}

var MessageType = struct {
	CreateRemoteTunnel string
	RemoteTunnelReady  string
	CloseRemoteTunnel  string

	CreateLocalTunnel string
	LocalTunnelReady  string
	CloseLocalTunnel  string

	Data  string
	Error string
}{
	CreateLocalTunnel: "CreateLocalTunnel",
	LocalTunnelReady:  "LocalTunnelReady",
	CloseLocalTunnel:  "CloseLocalTunnel",

	CreateRemoteTunnel: "CreateRemoteTunnel",
	RemoteTunnelReady:  "RemoteTunnelReady",
	CloseRemoteTunnel:  "CloseRemoteTunnel",

	Error: "Error",
	Data:  "Data",
}

func CreateLocalTunnelMessage(protocol string, service string) *Message {
	return &Message{
		Type:     MessageType.CreateLocalTunnel,
		Service:  service,
		Protocol: protocol,
	}
}

func CreateRemoteTunnelMessage(protocol string, service string) *Message {
	return &Message{
		Type:     MessageType.CreateRemoteTunnel,
		Service:  service,
		Protocol: protocol,
	}
}

func RemoteTunnelReadyMessage(protocol string, service string) *Message {
	return &Message{
		Type:     MessageType.RemoteTunnelReady,
		Service:  service,
		Protocol: protocol,
	}
}

func LocalTunnelReadyMessage(protocol string, service string) *Message {
	return &Message{
		Type:     MessageType.LocalTunnelReady,
		Service:  service,
		Protocol: protocol,
	}
}

func ErrorMessage(err error) *Message {
	return &Message{
		Type:        MessageType.Error,
		Description: fmt.Sprint(err),
	}
}

func DataMessage(clientId string, data []byte) *Message {
	return &Message{
		Type:     MessageType.Data,
		Data:     data,
		ClientId: clientId,
	}
}

type Message struct {
	Type        string `json:"t"`
	Description string `json:"d,omitempty"`
	Service     string `json:"s,omitempty"`
	Protocol    string `json:"p,omitempty"`
	ClientId    string `json:"c,omitempty"`
	Data        []byte `json:"b,omitempty"`
}
