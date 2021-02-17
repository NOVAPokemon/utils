package comms_manager

import (
	"fmt"
	"strconv"
	"sync"
	"time"

	"github.com/NOVAPokemon/utils/websockets"
	http "github.com/bruno-anjos/archimedesHTTPClient"
	"github.com/golang/geo/s2"
	"github.com/gorilla/websocket"
	"github.com/mitchellh/mapstructure"
	log "github.com/sirupsen/logrus"
)

type S2DelayedCommsManager struct {
	CellId       s2.CellID
	DelaysMatrix *DelaysMatrixType
	ClientDelays *ClientDelays
	websockets.CommsManagerWithCounter
	CommsManagerWithClient
	sync.RWMutex
}

var cellsToRegion = map[s2.CellID]string{
	s2.CellIDFromToken("54"): "ca-central-1",
	s2.CellIDFromToken("4c"): "eu-west-1",
	s2.CellIDFromToken("44"): "eu-central-1",
	s2.CellIDFromToken("5c"): "ap-northeast-2",
	s2.CellIDFromToken("7c"): "us-west-2",
	s2.CellIDFromToken("84"): "us-west-1",
	s2.CellIDFromToken("8c"): "us-east-1",
	s2.CellIDFromToken("0c"): "eu-west-3",
	s2.CellIDFromToken("14"): "eu-south-1",
	s2.CellIDFromToken("3c"): "me-south-1",
	s2.CellIDFromToken("34"): "ap-east-1",
	s2.CellIDFromToken("64"): "ap-northeast-1",
	s2.CellIDFromToken("74"): "ap-southeast-2",
	s2.CellIDFromToken("9c"): "sa-east-1",
	s2.CellIDFromToken("94"): "us-west-1",
	s2.CellIDFromToken("04"): "sa-east-1",
	s2.CellIDFromToken("1c"): "af-south-1",
	s2.CellIDFromToken("24"): "ap-south-1",
	s2.CellIDFromToken("2c"): "ap-southeast-1",
	s2.CellIDFromToken("6c"): "ap-southeast-2",
	s2.CellIDFromToken("a4"): "sa-east-1",
	s2.CellIDFromToken("bc"): "sa-east-1",
	s2.CellIDFromToken("b4"): "af-south-1",
	s2.CellIDFromToken("ac"): "ap-southeast-2",
}

func (d *S2DelayedCommsManager) ApplyReceiveLogic(msg *websockets.WebsocketMsg) *websockets.WebsocketMsg {
	if msg.MsgType != websocket.TextMessage || msg.Content.MsgKind != websockets.Wrapper {
		return msg
	} else if msg.Content.AppMsgType != websockets.Tagged {
		panic(fmt.Sprintf("s2delayed comms manager does not know how to treat %s", msg.Content.AppMsgType))
	}

	taggedMessage := &websockets.TaggedMessage{}
	if err := mapstructure.Decode(msg.Content.Data, taggedMessage); err != nil {
		panic(err)
	}

	cellId := s2.CellIDFromToken(taggedMessage.LocationTag)
	myRegionTag, requesterRegionTag, delay := d.getDelay(cellId, taggedMessage.IsClient)

	log.Infof("i am at %s got ws message from %s sleeping %f (isClient: %t)", myRegionTag, requesterRegionTag,
		delay, taggedMessage.IsClient)

	sleepDuration := time.Duration(delay) * time.Millisecond
	time.Sleep(sleepDuration)

	msg.Content = &taggedMessage.Content

	return msg
}

func (d *S2DelayedCommsManager) ApplySendLogic(msg *websockets.WebsocketMsg) *websockets.WebsocketMsg {
	if msg.MsgType == websocket.TextMessage {
		d.RLock()
		taggedMsg := websockets.TaggedMessage{
			LocationTag: d.CellId.ToToken(),
			IsClient:    d.IsClient,
			Content:     *msg.Content,
		}
		d.RUnlock()
		wrapperGenericMsg := taggedMsg.ConvertToWSMessage()
		return wrapperGenericMsg
	}

	return msg
}

func (d *S2DelayedCommsManager) WriteGenericMessageToConn(conn *websocket.Conn, msg *websockets.WebsocketMsg) error {
	d.DefaultCommsManager.ApplySendLogic(msg)
	msg = d.ApplySendLogic(msg)

	if msg.Content == nil {
		return conn.WriteMessage(msg.MsgType, nil)
	}

	return conn.WriteMessage(msg.MsgType, msg.Content.Serialize())
}

func (d *S2DelayedCommsManager) ReadMessageFromConn(conn *websocket.Conn) (*websockets.WebsocketMsg, error) {
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

func (d *S2DelayedCommsManager) DoHTTPRequest(client *http.Client, req *http.Request) (*http.Response, error) {
	req.Header.Set(locationTagKey, d.CellId.ToToken())
	req.Header.Set(tagIsClientKey, strconv.FormatBool(d.IsClient))

	var (
		resp *http.Response
		err  error
	)

	for {
		resp, err = client.Do(req)
		if d.CommsManagerWithCounter.LogRequestAndRetry(err) {
			break
		}

		<-time.After(timeBetweenRetries)
	}

	return resp, err
}

func (d *S2DelayedCommsManager) HTTPRequestInterceptor(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestLocationToken := r.Header.Get(locationTagKey)
		if requestLocationToken == "" {
			next.ServeHTTP(w, r)
			return
		}
		requesterIsClient, err := strconv.ParseBool(r.Header.Get(tagIsClientKey))
		if err != nil {
			panic(fmt.Sprintf("could not parse %+v to bool", requesterIsClient))
		}

		myRegionTag, requesterRegionTag, delay := d.getDelay(s2.CellIDFromToken(requestLocationToken),
			requesterIsClient)

		log.Infof("i am at %s got request from %s sleeping %f", myRegionTag, requesterRegionTag, delay)

		sleepDuration := time.Duration(delay) * time.Millisecond
		time.Sleep(sleepDuration)
		next.ServeHTTP(w, r)
	})
}

func (d *S2DelayedCommsManager) getDelay(requesterCell s2.CellID, isClient bool) (myRegionTag,
	requesterRegionTag string, delay float64) {
	d.RLock()
	myRegionTag = TranslateCellToRegion(d.CellId)
	d.RUnlock()
	requesterRegionTag = TranslateCellToRegion(requesterCell)

	var ok bool
	delay, ok = (*d.DelaysMatrix)[requesterRegionTag][myRegionTag]
	if !ok {
		panic(fmt.Sprintf("could not delay WS message from %s to %s", requesterRegionTag, myRegionTag))
	}

	return
}

func TranslateCellToRegion(c s2.CellID) (locationTag string) {
	var ok bool
	locationTag, ok = cellsToRegion[c.Parent(1)]
	if !ok {
		log.Fatalf("no region tag for cell %s", c.Parent(1).ToToken())
	}

	return
}

func (d *S2DelayedCommsManager) SetCellID(cellID s2.CellID) {
	d.Lock()
	d.CellId = cellID
	d.Unlock()
}
