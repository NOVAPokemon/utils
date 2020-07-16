package websockets

import (
	"net/http"

	"github.com/gorilla/websocket"
)

type CommunicationManager interface {
	WriteGenericMessageToConn(conn *websocket.Conn, msg GenericMsg) error
	ReadTextMessageFromConn(conn *websocket.Conn) ([]byte, error)
	DoHTTPRequest(client *http.Client, req *http.Request) (*http.Response, error)
	HTTPRequestInterceptor(next http.Handler) http.Handler
}
