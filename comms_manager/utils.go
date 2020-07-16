package comms_manager

import (
	"net/http"

	"github.com/NOVAPokemon/utils/websockets"
	"github.com/gorilla/websocket"
)

type CommunicationManager interface {
	WriteTextMessageToConn(conn *websocket.Conn, serializable websockets.Serializable) error
	WriteNonTextMessageToConn(conn *websocket.Conn, msgType int, data []byte) error
	ReadTextMessageFromConn(conn *websocket.Conn) ([]byte, error)
	DoHTTPRequest(client *http.Client, req *http.Request) (*http.Response, error)
	HTTPRequestInterceptor(next http.Handler) http.Handler
}
