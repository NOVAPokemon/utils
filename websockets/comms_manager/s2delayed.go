package comms_manager

import (
	"encoding/json"
	"fmt"
	"os"
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
	websockets.CommsManagerWithCounter
	CommsManagerWithClient
	sync.RWMutex
}

const (
	cellsToRegionFilePath = "/service/cells_to_region.json"
	delayAppliedKey       = "Delay_applied"
)

var cellsToRegion = map[s2.CellID]string{}

func init() {
	f, err := os.Open(cellsToRegionFilePath)
	if err != nil {
		log.Panic(err)
	}

	cellTokensToRegion := map[string]string{}

	err = json.NewDecoder(f).Decode(&cellTokensToRegion)
	if err != nil {
		log.Panic(err)
	}

	log.Info("loading cells to region...")

	for cellToken, region := range cellTokensToRegion {
		cellsToRegion[s2.CellIDFromToken(cellToken)] = region
	}
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

	splitDelay := delay / 2
	log.Infof("i am at %s got ws message from %s sleeping %f (isClient: %t)", myRegionTag, requesterRegionTag,
		splitDelay, taggedMessage.IsClient)

	sleepDuration := time.Duration(splitDelay) * time.Millisecond
	time.Sleep(sleepDuration)

	msg.Content = &taggedMessage.Content

	return msg
}

func (d *S2DelayedCommsManager) ApplySendLogic(msg *websockets.WebsocketMsg) *websockets.WebsocketMsg {
	if msg.MsgType == websocket.TextMessage {
		d.RLock()
		taggedMsg := websockets.TaggedMessage{
			LocationTag: d.GetCellID().ToToken(),
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
	req.Header.Set(LocationTagKey, d.GetCellID().ToToken())
	req.Header.Set(TagIsClientKey, strconv.FormatBool(d.IsClient))

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

	if resp != nil && resp.Header != nil {
		if responderLocationToken := resp.Header.Get(serverLocationTagKey); responderLocationToken != "" {
			delayString := resp.Header.Get(delayAppliedKey)

			var delay float64
			delay, err = strconv.ParseFloat(delayString, 64)

			if err != nil {
				log.Panic(resp, err)
			}

			log.Infof("applying %f delay", delay)

			sleepDuration := time.Duration(delay) * time.Millisecond
			time.Sleep(sleepDuration)
		} else {
			log.Info("no server location tag")
		}
	} else {
		log.Infof("something was nil: %+v", resp)
	}

	return resp, err
}

func (d *S2DelayedCommsManager) HTTPRequestInterceptor(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestLocationToken := r.Header.Get(LocationTagKey)
		if requestLocationToken == "" {
			log.Infof("request %+v did not have a location tag", r)

			next.ServeHTTP(w, r)
			return
		}

		w.Header().Set(serverLocationTagKey, d.GetCellID().ToToken())

		requesterIsClient, err := strconv.ParseBool(r.Header.Get(TagIsClientKey))
		if err != nil {
			log.Panic(fmt.Sprintf("could not parse %+v to bool", requesterIsClient))
		}

		myRegionTag, requesterRegionTag, delay := d.getDelay(s2.CellIDFromToken(requestLocationToken),
			requesterIsClient)
		splitDelay := delay / 2

		w.Header().Set(delayAppliedKey, fmt.Sprintf("%f", splitDelay))

		log.Infof("i am at %s got request from %s sleeping %f", myRegionTag, requesterRegionTag, splitDelay)

		sleepDuration := time.Duration(splitDelay) * time.Millisecond
		time.Sleep(sleepDuration)

		next.ServeHTTP(w, r)
	})
}

func (d *S2DelayedCommsManager) getDelay(requesterCell s2.CellID, isClient bool) (myRegionTag,
	requesterRegionTag string, delay float64) {
	d.RLock()
	myRegionTag = TranslateCellToRegion(d.GetCellID())
	d.RUnlock()
	requesterRegionTag = TranslateCellToRegion(requesterCell)

	if isClient {
		delay = 2 * (*d.DelaysMatrix)[requesterRegionTag][requesterRegionTag]
		log.Infof("adding %f ms from client to node", delay)

		if requesterRegionTag != myRegionTag {
			delay += (*d.DelaysMatrix)[requesterRegionTag][myRegionTag]
		}
	} else {
		var ok bool
		delay, ok = (*d.DelaysMatrix)[requesterRegionTag][myRegionTag]
		if !ok {
			panic(fmt.Sprintf("could not delay WS message from %s to %s", requesterRegionTag, myRegionTag))
		}
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

func (d *S2DelayedCommsManager) GetCellID() s2.CellID {
	d.RLock()
	cellID := d.CellId
	d.RUnlock()

	return cellID
}
