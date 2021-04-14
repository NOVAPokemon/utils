package websockets

import (
	"strings"
	"sync/atomic"

	http "github.com/bruno-anjos/archimedesHTTPClient"

	"github.com/gorilla/websocket"
	log "github.com/sirupsen/logrus"
)

type CommunicationManager interface {
	ApplyReceiveLogic(msg *WebsocketMsg) *WebsocketMsg
	ApplySendLogic(msg *WebsocketMsg) *WebsocketMsg
	WriteGenericMessageToConn(conn *websocket.Conn, msg *WebsocketMsg) error
	ReadMessageFromConn(conn *websocket.Conn) (<-chan *WebsocketMsg, error)
	DoHTTPRequest(client *http.Client, req *http.Request) (*http.Response, error)
	HTTPRequestInterceptor(next http.Handler) http.Handler
}

type CommsManagerWithCounter struct {
	RetriesCount  int64
	RequestsCount int64
}

func (d *CommsManagerWithCounter) LogRequestAndRetry(err error, ts int64) (success bool) {
	success = true

	log.Infof("[REQ] %d %d", ts, atomic.AddInt64(&d.RequestsCount, 1))

	if err != nil {
		log.Warnf("[REQ_ERR] %d %s", ts, err.Error())
		success = false
		log.Infof("[RET] %d %d", ts, atomic.AddInt64(&d.RetriesCount, 1))
	}

	return
}

const (
	errConnRefused = "connection refused"
	errTimeout1    = "timeout"
	errTimeout2    = "Timeout"
	errBodyLength  = "with Body length"
)

func checkErr(err error) bool {
	if err == nil {
		return false
	}

	if strings.Contains(err.Error(), errConnRefused) || strings.Contains(err.Error(), errTimeout1) ||
		strings.Contains(err.Error(), errTimeout2) || strings.Contains(err.Error(), errBodyLength) {
		return true
	}

	return false
}
