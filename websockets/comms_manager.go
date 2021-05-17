package websockets

import (
	"fmt"
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

func (d *CommsManagerWithCounter) LogRequestAndRetry(resp *http.Response, err error, ts int64) (success bool) {
	success = true

	statusCodeIsTimeout := false
	if resp != nil {
		statusCodeIsTimeout = resp.StatusCode == http.StatusRequestTimeout || resp.StatusCode == http.StatusGatewayTimeout
	}

	// statusCodeIsTimeout is only true when resp is != than nil, therefore not sure why nilness analyzer fails
	if statusCodeIsTimeout {
		err = fmt.Errorf("got timeout status code %d", resp.StatusCode)
	}

	log.Infof("[REQ] %d %d", ts, atomic.AddInt64(&d.RequestsCount, 1))

	if err != nil {
		log.Warnf("[REQ_ERR] %d %s", ts, err.Error())
		success = false
		log.Infof("[RET] %d %d", ts, atomic.AddInt64(&d.RetriesCount, 1))
	}

	return
}

const (
	ErrConnRefused = "connection refused"
	ErrTimeout     = "timeout"
	errTimeout2    = "Timeout"
	errBodyLength  = "with Body length"
)

func checkErr(err error) bool {
	if err == nil {
		return false
	}

	if strings.Contains(err.Error(), ErrConnRefused) || strings.Contains(err.Error(), ErrTimeout) ||
		strings.Contains(err.Error(), errTimeout2) || strings.Contains(err.Error(), errBodyLength) {
		return true
	}

	return false
}
