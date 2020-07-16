package comms_manager

import (
	"net/http"

	"github.com/NOVAPokemon/utils/websockets"
	"github.com/gorilla/websocket"
)

type DefaultCommsManager struct{}

func (d *DefaultCommsManager) WriteGenericMessageToConn(conn *websocket.Conn, msg websockets.GenericMsg) error {
	return conn.WriteMessage(msg.MsgType, msg.Data)
}

func (d *DefaultCommsManager) ReadTextMessageFromConn(conn *websocket.Conn) ([]byte, error) {
	_, p, err := conn.ReadMessage()
	return p, err
}

func (d *DefaultCommsManager) DoHTTPRequest(client *http.Client, req *http.Request) (*http.Response, error) {
	return client.Do(req)
}

func (d *DefaultCommsManager) HTTPRequestInterceptor(next http.Handler) http.Handler {
	return next
}
