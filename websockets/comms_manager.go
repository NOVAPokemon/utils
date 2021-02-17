package websockets

import (
	"net"
	"sync/atomic"
	"syscall"

	http "github.com/bruno-anjos/archimedesHTTPClient"

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

func checkErr(err error) bool {
	if err == nil {
		return false
	}

	if netError, ok := err.(net.Error); ok && netError.Timeout() {
		return true
	}

	switch t := err.(type) {
	case *net.OpError:
		if t.Op == "dial" {
			log.Panic(err.Error())
		} else if t.Op == "read" {
			return true
		}
	case syscall.Errno:
		if t == syscall.ECONNREFUSED {
			return true
		}
	default:
		log.Panic(err.Error())
	}

	return false
}
