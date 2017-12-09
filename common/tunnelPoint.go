package common

type TunnelPoint interface {
	ReceiveDataFromClientSocket(client *Client, data []byte)
	CloseClient(clientId string)
	IsOpen() bool
}
