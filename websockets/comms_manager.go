package websockets

import (
	"net/http"

	"github.com/gorilla/websocket"
)

type CommunicationManager interface {
	WriteGenericMessageToConn(conn *websocket.Conn, msg GenericMsg) error
	ReadMessageFromConn(conn *websocket.Conn) (int, []byte, error)
	DoHTTPRequest(client *http.Client, req *http.Request) (*http.Response, error)
	HTTPRequestInterceptor(next http.Handler) http.Handler
}
