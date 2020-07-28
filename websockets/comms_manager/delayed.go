package comms_manager

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/NOVAPokemon/utils/websockets"
	"github.com/gorilla/websocket"
	"github.com/mitchellh/mapstructure"
	log "github.com/sirupsen/logrus"
)

const (
	locationTagKey = "Location_tag"
	tagIsClientKey = "Tag_is_client"
)

type (
	DelaysMatrixType = map[string]map[string]float64

	ClientDelays struct {
		Default     float64            `json:"default"`
		Multipliers map[string]float64 `json:"multipliers"`
	}
)

type DelayedCommsManager struct {
	LocationTag  string
	DelaysMatrix *DelaysMatrixType
	ClientDelays *ClientDelays
	CommsManagerWithClient
}

func (d *DelayedCommsManager) ApplyReceiveLogic(msg *websockets.WebsocketMsg) *websockets.WebsocketMsg {
	if msg.MsgType != websocket.TextMessage || msg.Content.MsgKind != websockets.Wrapper {
		return msg
	} else if msg.Content.AppMsgType != websockets.Tagged {
		panic(fmt.Sprintf("delayed comms manager does not know how to treat %s", msg.Content.AppMsgType))
	}

	log.Infof("Msg: %+v", msg)
	log.Infof("Content: %+v", msg.Content)

	taggedMessage := &websockets.TaggedMessage{}
	if err := mapstructure.Decode(msg.Content.Data, taggedMessage); err != nil {
		panic(err)
	}
	log.Infof("Converted: %+v", taggedMessage)

	requesterLocationTag := taggedMessage.LocationTag
	delay := d.getDelay(requesterLocationTag, taggedMessage.IsClient)

	sleepDuration := time.Duration(delay) * time.Millisecond
	time.Sleep(sleepDuration)

	msg.Content = &taggedMessage.Content

	return msg
}

func (d *DelayedCommsManager) ApplySendLogic(msg *websockets.WebsocketMsg) *websockets.WebsocketMsg {
	if msg.MsgType == websocket.TextMessage {
		taggedMsg := websockets.TaggedMessage{
			LocationTag: d.LocationTag,
			IsClient:    d.IsClient,
			Content:     *msg.Content,
		}
		wrapperGenericMsg := taggedMsg.ConvertToWSMessage()
		log.Info("Will send: %+v", wrapperGenericMsg)
		return wrapperGenericMsg
	}

	return msg
}

func (d *DelayedCommsManager) WriteGenericMessageToConn(conn *websocket.Conn, msg *websockets.WebsocketMsg) error {
	d.DefaultCommsManager.ApplySendLogic(msg)
	msg = d.ApplySendLogic(msg)

	if msg.Content == nil {
		return conn.WriteMessage(msg.MsgType, []byte{})
	}

	return conn.WriteMessage(msg.MsgType, msg.Content.Serialize())
}

func (d *DelayedCommsManager) ReadMessageFromConn(conn *websocket.Conn) (*websockets.WebsocketMsg, error) {
	msgType, p, err := conn.ReadMessage()
	if err != nil {
		log.Warn(err)
		return nil, err
	}

	taggedContent := websockets.ParseContent(p)
	wsMsg := &websockets.WebsocketMsg{
		MsgType: msgType,
		Content: taggedContent,
	}
	msg := d.ApplyReceiveLogic(wsMsg)

	d.DefaultCommsManager.ApplyReceiveLogic(msg)

	return msg, nil
}

func (d *DelayedCommsManager) DoHTTPRequest(client *http.Client, req *http.Request) (*http.Response, error) {
	req.Header.Set(locationTagKey, d.LocationTag)
	req.Header.Set(tagIsClientKey, strconv.FormatBool(d.IsClient))
	return client.Do(req)
}

func (d *DelayedCommsManager) HTTPRequestInterceptor(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requesterLocationTag := r.Header.Get(locationTagKey)
		if requesterLocationTag == "" {
			next.ServeHTTP(w, r)
			return
		}
		requesterIsClient, err := strconv.ParseBool(r.Header.Get(tagIsClientKey))
		if err != nil {
			panic(fmt.Sprintf("could not parse %+v to bool", requesterIsClient))
		}

		delay := d.getDelay(requesterLocationTag, requesterIsClient)

		sleepDuration := time.Duration(delay) * time.Millisecond
		time.Sleep(sleepDuration)
		next.ServeHTTP(w, r)
	})
}

func (d *DelayedCommsManager) getDelay(requesterLocationTag string, isClient bool) (delay float64) {
	var ok bool
	if !isClient {
		delay, ok = (*d.DelaysMatrix)[requesterLocationTag][d.LocationTag]
		if !ok {
			panic(fmt.Sprintf("could not delay WS message from %s to %s", requesterLocationTag, d.LocationTag))
		}
	} else {
		var multiplier float64
		multiplier, ok = (*d.ClientDelays).Multipliers[d.LocationTag]
		if !ok {
			panic(fmt.Sprintf("could not delay WS message from client %s to %s",
				requesterLocationTag, d.LocationTag))
		}

		delay = (*d.ClientDelays).Default * multiplier
	}

	return delay
}
