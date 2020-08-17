package websockets

import (
	"encoding/json"
	"time"

	http "github.com/bruno-anjos/archimedesHTTPClient"

	"github.com/gorilla/websocket"
	log "github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type MsgKinds = int

const (
	Request       = MsgKinds(0)
	Reply         = MsgKinds(1)
	Wrapper       = MsgKinds(2)
	NotApplicable = MsgKinds(3)
)

type WebsocketMsgContent struct {
	AppMsgType   string
	Data         interface{}
	MsgKind      MsgKinds
	RequestTrack *TrackedInfo
}

func (msg WebsocketMsgContent) Serialize() []byte {
	jsonbytes, err := json.Marshal(msg)
	if err != nil {
		panic(err)
	}
	return jsonbytes
}

type WebsocketMsg struct {
	MsgType int
	Content *WebsocketMsgContent
}

func NewRequestMsg(appMsgType string, data interface{}) *WebsocketMsg {
	return &WebsocketMsg{
		MsgType: websocket.TextMessage,
		Content: &WebsocketMsgContent{
			AppMsgType:   appMsgType,
			Data:         data,
			MsgKind:      Request,
			RequestTrack: NewTrackedInfo(primitive.NewObjectID()),
		},
	}
}

func NewReplyMsg(appMsgType string, content interface{}, requestTrack TrackedInfo) *WebsocketMsg {
	return &WebsocketMsg{
		MsgType: websocket.TextMessage,
		Content: &WebsocketMsgContent{
			AppMsgType:   appMsgType,
			Data:         content,
			MsgKind:      Reply,
			RequestTrack: &requestTrack,
		},
	}
}

func NewWrapperMsg(wrapperMsgType string, content interface{}) *WebsocketMsg {
	return &WebsocketMsg{
		MsgType: websocket.TextMessage,
		Content: &WebsocketMsgContent{
			AppMsgType:   wrapperMsgType,
			Data:         content,
			MsgKind:      Wrapper,
			RequestTrack: nil,
		},
	}
}

func NewStandardMsg(appMsgType string, content interface{}) *WebsocketMsg {
	return &WebsocketMsg{
		MsgType: websocket.TextMessage,
		Content: &WebsocketMsgContent{
			AppMsgType:   appMsgType,
			Data:         content,
			MsgKind:      NotApplicable,
			RequestTrack: nil,
		},
	}
}

func NewControlMsg(wsMsgType int) *WebsocketMsg {
	return &WebsocketMsg{
		MsgType: wsMsgType,
		Content: nil,
	}
}

type SerializableToWS interface {
	ConvertToWSMessage() *WebsocketMsg
}

type ReplySerializableToWS interface {
	ConvertToWSMessage(info TrackedInfo) *WebsocketMsg
}

const Tagged = "TAGGED"

type TaggedMessage struct {
	LocationTag string
	IsClient    bool
	Content     WebsocketMsgContent
}

func (tMsg TaggedMessage) ConvertToWSMessage() *WebsocketMsg {
	return NewWrapperMsg(Tagged, tMsg)
}

type Trackable interface {
	Emit(emitTimestamp int64)
	Receive(receiveTimestamp int64)
	LogEmit()
	LogReceive()
}

const (
	invalidTime         = -1
	TrackInfoHeaderName = "Track_info"
)

type TrackedInfo struct {
	TimeEmitted  int64
	TimeReceived int64
	Id           string
}

func NewTrackedInfo(id primitive.ObjectID) *TrackedInfo {
	return &TrackedInfo{
		TimeEmitted:  invalidTime,
		TimeReceived: invalidTime,
		Id:           id.Hex(),
	}
}

func (ti *TrackedInfo) Emit(emitTimestamp int64) {
	ti.TimeEmitted = emitTimestamp
}

func (ti *TrackedInfo) Receive(receiveTimestamp int64) {
	ti.TimeReceived = receiveTimestamp
}

func (ti *TrackedInfo) LogEmit(msgType string) {
	if ti.TimeEmitted == invalidTime {
		log.Error(ErrorTooEarlyToLogEmit)
		return
	}

	log.Infof("[EMIT] %s %s %d", msgType, ti.Id, ti.TimeEmitted)
}

func (ti *TrackedInfo) LogReceive(msgType string) {
	if ti.TimeReceived == invalidTime {
		log.Error(ErrorTooEarlyToLogReceive)
	}

	log.Infof("[RECEIVE] %s %s %d", msgType, ti.Id, ti.TimeReceived)
}

func (ti *TrackedInfo) TimeTook() (int64, bool) {
	if ti.TimeEmitted == invalidTime {
		log.Error(ErrorMsgWasNotEmmitted)
		return invalidTime, false
	}

	if ti.TimeReceived == invalidTime {
		log.Error(ErrorMsgWasNotEmmitted)
		return invalidTime, false
	}

	return ti.TimeReceived - ti.TimeEmitted, true
}

func (ti *TrackedInfo) SerializeToJSON() string {
	jsonbytes, err := json.Marshal(ti)
	if err != nil {
		panic(err)
	}
	return string(jsonbytes)
}

func DeserializeTrackInfo(jsonMsg string) TrackedInfo {
	var info TrackedInfo
	err := json.Unmarshal([]byte(jsonMsg), &info)
	if err != nil {
		panic(err)
	}

	return info
}

func AddTrackInfoToHeader(h *http.Header, msgType string) {
	trackInfo := NewTrackedInfo(primitive.NewObjectID())
	trackInfo.Emit(MakeTimestamp())
	trackInfo.LogEmit(msgType)

	h.Set(TrackInfoHeaderName, trackInfo.SerializeToJSON())
}

func GetTrackInfoFromHeader(h *http.Header) TrackedInfo {
	trackedInfoJSON := h.Get(TrackInfoHeaderName)
	return DeserializeTrackInfo(trackedInfoJSON)
}

func MakeTimestamp() int64 {
	return time.Now().UnixNano() / (int64(time.Millisecond) / int64(time.Nanosecond))
}
