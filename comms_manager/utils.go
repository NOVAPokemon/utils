package comms_manager

import (
	"net/http"

	"github.com/gorilla/websocket"
)

type WsConnWithServerName struct {
	websocket.Conn
	ServerName string
}

type CommunicationManager interface {
	WriteMessageToConn(conn *WsConnWithServerName, msgType int, data []byte) error
	DoHTTPRequest(client *http.Client, req *http.Request) (*http.Response, error)
}
