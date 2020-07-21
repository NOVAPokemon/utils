package comms_manager

import (
	"net/http"

	"github.com/NOVAPokemon/utils/websockets"
	"github.com/gorilla/websocket"
)

type CommsManagerWithClient struct {
	IsClient bool
}

type DefaultCommsManager struct{}

func (d *DefaultCommsManager) WriteGenericMessageToConn(conn *websocket.Conn, msg websockets.GenericMsg) error {
	return conn.WriteMessage(msg.MsgType, msg.Data)
}

func (d *DefaultCommsManager) ReadMessageFromConn(conn *websocket.Conn) (int, []byte, error) {
	return conn.ReadMessage()
}

func (d *DefaultCommsManager) DoHTTPRequest(client *http.Client, req *http.Request) (*http.Response, error) {
	return client.Do(req)
}

func (d *DefaultCommsManager) HTTPRequestInterceptor(next http.Handler) http.Handler {
	return next
}
