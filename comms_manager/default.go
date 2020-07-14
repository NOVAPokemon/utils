package comms_manager

import (
	"net/http"
)

type DefaultCommsManager struct{}

func (w *DefaultCommsManager) WriteMessageToConn(conn *WsConnWithServerName, msgType int, data []byte) error {
	return conn.WriteMessage(msgType, data)
}

func (w *DefaultCommsManager) DoHTTPRequest(client *http.Client, req *http.Request) (*http.Response, error) {
	return client.Do(req)
}
