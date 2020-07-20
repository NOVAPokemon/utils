package comms_manager

import (
	"net/http"
	"time"

	"github.com/NOVAPokemon/utils/websockets"
	"github.com/gorilla/websocket"
	log "github.com/sirupsen/logrus"
)

const (
	locationTagKey = "Location_tag"
)

type DelaysMatrixType = map[string]map[string]float64

type DelayedCommsManager struct {
	LocationTag  string
	DelaysMatrix *DelaysMatrixType
}

func (d *DelayedCommsManager) WriteGenericMessageToConn(conn *websocket.Conn, msg websockets.GenericMsg) error {
	if msg.MsgType == websocket.TextMessage {
		taggedMsg := websockets.TaggedMessage{
			LocationTag: d.LocationTag,
			MsgBytes:    msg.Data,
		}.SerializeToWSMessage().Serialize()

		return conn.WriteMessage(websocket.TextMessage, []byte(taggedMsg))
	}
	return conn.WriteMessage(msg.MsgType, msg.Data)
}

func (d *DelayedCommsManager) ReadMessageFromConn(conn *websocket.Conn) (int, []byte, error) {
	msgType, p, err := conn.ReadMessage()
	if err != nil {
		return 0, nil, err
	}

	log.Infof("deserializing %s", string(p))
	taggedMessage, err := websockets.DeserializeTaggedMessage(p)
	if err != nil {
		panic(err)
	}

	log.Infof("result %+v", taggedMessage)

	requesterLocationTag := taggedMessage.LocationTag
	sleepDuration := time.Duration((*d.DelaysMatrix)[requesterLocationTag][d.LocationTag]) * time.Millisecond
	time.Sleep(sleepDuration)

	return msgType, taggedMessage.MsgBytes, nil
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
