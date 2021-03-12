package websockets

import (
	"net/http"
	"strings"
	"sync/atomic"

	"github.com/gorilla/websocket"
	log "github.com/sirupsen/logrus"
)

type CommunicationManager interface {
	ApplyReceiveLogic(msg *WebsocketMsg) *WebsocketMsg
	ApplySendLogic(msg *WebsocketMsg) *WebsocketMsg
	WriteGenericMessageToConn(conn *websocket.Conn, msg *WebsocketMsg) error
	ReadMessageFromConn(conn *websocket.Conn) (*WebsocketMsg, error)
	DoHTTPRequest(client *http.Client, req *http.Request) (*http.Response, error)
	HTTPRequestInterceptor(next http.Handler) http.Handler
}

type CommsManagerWithCounter struct {
	RetriesCount  int64
	RequestsCount int64
}

func (d *CommsManagerWithCounter) LogRequestAndRetry(err error) (success bool) {
	success = true
	failed := false

	log.Infof("Requests count: %d", atomic.AddInt64(&d.RequestsCount, 1))

	if hasErr := checkErr(err); hasErr {
		success = false
		failed = true
	}

	if failed {
		log.Infof("Retries count: %d", atomic.AddInt64(&d.RetriesCount, 1))
	}

	return
}

const (
	errConnRefused = "connection refused"
	errTimeout1    = "timeout"
	errTimeout2    = "Timeout"
)

func checkErr(err error) bool {
	if err == nil {
		return false
	}

	if strings.Contains(err.Error(), errConnRefused) || strings.Contains(err.Error(), errTimeout1) ||
		strings.Contains(err.Error(), errTimeout2) {
		return true
	}

	return false
}
