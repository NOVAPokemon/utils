package comms_manager

import (
	"net/http"
	"time"

	"github.com/NOVAPokemon/utils/websockets"
	"github.com/gorilla/websocket"
)

const (
	locationTagKey = "Location_tag"
)

type DelaysMatrixType = map[string]map[string]int

type DelayedCommsManager struct {
	LocationTag  string
	DelaysMatrix *DelaysMatrixType
}

func (d *DelayedCommsManager) WriteTextMessageToConn(conn *websocket.Conn,
	serializable websockets.Serializable) error {

	msg := websockets.TaggedMessage{
		LocationTag: d.LocationTag,
		MsgBytes:    []byte(serializable.SerializeToWSMessage().Serialize()),
	}.SerializeToWSMessage().Serialize()

	return conn.WriteMessage(websocket.TextMessage, []byte(msg))
}

func (d *DelayedCommsManager) WriteNonTextMessageToConn(conn *websocket.Conn, msgType int, data []byte) error {
	return conn.WriteMessage(msgType, data)
}

func (d *DelayedCommsManager) ReadTextMessageFromConn(conn *websocket.Conn) ([]byte, error) {
	_, p, err := conn.ReadMessage()
	if err != nil {
		return nil, err
	}

	taggedMessage, err := websockets.DeserializeTaggedMessage(p)
	if err != nil {
		panic(err)
	}

	requesterLocationTag := taggedMessage.LocationTag
	sleepDuration := time.Duration((*d.DelaysMatrix)[requesterLocationTag][d.LocationTag]) * time.Millisecond
	time.Sleep(sleepDuration)

	return taggedMessage.MsgBytes, nil
}

func (d *DelayedCommsManager) DoHTTPRequest(client *http.Client, req *http.Request) (*http.Response, error) {
	req.Header.Set(locationTagKey, d.LocationTag)
	return client.Do(req)
}

func (d *DelayedCommsManager) HTTPRequestInterceptor(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requesterLocationTag := r.Header.Get(locationTagKey)
		sleepDuration := time.Duration((*d.DelaysMatrix)[requesterLocationTag][d.LocationTag]) * time.Millisecond
		time.Sleep(sleepDuration)
		next.ServeHTTP(w, r)
	})
}
