package comms_manager

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/NOVAPokemon/utils/websockets"
	"github.com/golang/geo/s1"
	"github.com/golang/geo/s2"
	"github.com/gorilla/websocket"
	"github.com/mitchellh/mapstructure"
	log "github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type S2DelayedCommsManager struct {
	CellId       s2.CellID
	DelaysMatrix *DelaysMatrixType
	websockets.CommsManagerWithCounter
	CommsManagerWithClient
	sync.RWMutex

	MyClosestNode int
}

const (
	delayAppliedKey = "Delay_applied"
	RequestIDKey    = "Request_id"
)

var (
	cellsToRegionFilePath string

	cellsToRegion = map[s2.CellID]string{}
	latencies     map[int][]float64
)

func init() {
	var ok bool
	if cellsToRegionFilePath, ok = os.LookupEnv("CELLS_TO_REGION"); !ok {
		cellsToRegionFilePath = "/service/cells_to_region.json"
	}

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

	log.Info("loading node latencies...")
	latencies = loadNodeLatencies()

	log.Info("Done!")
}

const (
	nodeNumEnvVar = "NODE_NUM"
)

func NewS2DelayedCommsManager(cellID s2.CellID, delaysConfig *DelaysMatrixType, isClient bool) *S2DelayedCommsManager {
	closestNode := -1
	if isClient {
		closestNode = getClosestNode(cellID)
	} else {
		nodeNumString, ok := os.LookupEnv(nodeNumEnvVar)
		if !ok {
			log.Panic("missing NODE_NUM env var")
		}

		var err error
		closestNode, err = strconv.Atoi(nodeNumString)
		if err != nil {
			log.Panic(err)
		}
	}

	manager := &S2DelayedCommsManager{
		CellId:                  cellID,
		DelaysMatrix:            delaysConfig,
		CommsManagerWithCounter: websockets.CommsManagerWithCounter{},
		CommsManagerWithClient: CommsManagerWithClient{
			IsClient: isClient,
		},
		MyClosestNode: closestNode,
		RWMutex:       sync.RWMutex{},
	}

	return manager
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

	myRegionTag, requesterRegionTag, delay := d.getWSDelay(cellId, taggedMessage.NodeNum)

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
			NodeNum:     d.MyClosestNode,
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

func (d *S2DelayedCommsManager) ReadMessageFromConn(conn *websocket.Conn) (<-chan *websockets.WebsocketMsg, error) {
	msgType, p, err := conn.ReadMessage()
	if err != nil {
		return nil, err
	}

	taggedContent := websockets.ParseContent(p)
	wsMsg := &websockets.WebsocketMsg{
		MsgType: msgType,
		Content: taggedContent,
	}

	msgChan := make(chan *websockets.WebsocketMsg)
	go func() {
		log.Infof("Before logic %+v", wsMsg)
		msg := d.ApplyReceiveLogic(wsMsg)
		log.Infof("After logic %+v", msg)
		d.DefaultCommsManager.ApplyReceiveLogic(msg)
		msgChan <- msg
	}()

	return msgChan, nil
}

func (d *S2DelayedCommsManager) DoHTTPRequest(client *http.Client, req *http.Request) (*http.Response, error) {
	req.Header.Set(LocationTagKey, d.GetCellID().ToToken())
	req.Header.Set(TagIsClientKey, strconv.FormatBool(d.IsClient))
	req.Header.Set(ClosestNodeKey, strconv.Itoa(d.MyClosestNode))

	var (
		resp *http.Response
		err  error
	)

	var bodyBytes []byte
	if req.Body != nil {
		bodyBytes, err = ioutil.ReadAll(req.Body)
		if err != nil {
			log.Panic(err)
		}
	}

	for {
		ts := websockets.MakeTimestamp()
		if d.IsClient {
			requestID := primitive.NewObjectID().Hex()
			req.Header.Set(RequestIDKey, requestID)
			log.Infof("[SENT_REQ_ID] %d %s", ts, requestID)
		}

		if req.Body != nil {
			req.Body = ioutil.NopCloser(bytes.NewBuffer(bodyBytes))
		}

		resp, err = client.Do(req)

		if d.IsClient && d.CommsManagerWithCounter.LogRequestAndRetry(resp, err,
			ts) {
			break
		}

		waitingTime := time.Duration(rand.Int31n(maxTimeBetweenRetries-minTimeBetweenRetries+1)+minTimeBetweenRetries) * time.Second
		<-time.After(waitingTime)
	}

	if resp != nil && resp.Header != nil {
		if d.IsClient {
			requestID := resp.Header.Get(RequestIDKey)
			log.Infof("[GOT_RESP_ID] %s", requestID)
		}

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
		if requestID := r.Header.Get(RequestIDKey); requestID != "" {
			ts := websockets.MakeTimestamp()
			log.Infof("[GOT_REQ_ID] %d %s", ts, requestID)
			w.Header().Set(RequestIDKey, requestID)
		}

		requestLocationToken := r.Header.Get(LocationTagKey)
		if requestLocationToken == "" {
			log.Infof("request %+v did not have a location tag", r)

			next.ServeHTTP(w, r)
			return
		}

		w.Header().Set(serverLocationTagKey, d.GetCellID().ToToken())

		requesterIsClient, err := strconv.ParseBool(r.Header.Get(TagIsClientKey))
		if err != nil {
			log.Panic("could not parse to bool")
		}

		closestNode, err := strconv.Atoi(r.Header.Get(ClosestNodeKey))
		if err != nil {
			log.Panic("could not parse to int %+v", closestNode)
		}

		myRegionTag, requesterRegionTag, delay := d.getHTTPDelay(s2.CellIDFromToken(requestLocationToken),
			requesterIsClient, closestNode)

		splitDelay := delay / 2

		w.Header().Set(delayAppliedKey, fmt.Sprintf("%f", splitDelay))

		log.Infof("i am at %s got request from %s sleeping %f", myRegionTag, requesterRegionTag, splitDelay)

		sleepDuration := time.Duration(splitDelay) * time.Millisecond
		time.Sleep(sleepDuration)
		next.ServeHTTP(w, r)
	})
}

func (d *S2DelayedCommsManager) getHTTPDelay(requesterCell s2.CellID, requesterIsClient bool,
	requesterClosestNode int) (myRegionTag, requesterRegionTag string, delay float64) {
	d.RLock()
	myRegionTag = TranslateCellToRegion(d.GetCellID())
	d.RUnlock()
	requesterRegionTag = TranslateCellToRegion(requesterCell)

	if requesterIsClient {
		// we will apply the delay from client -> closestNode and if closest node is not the one being used
		// add the delay from closestNode -> targetNode

		delay = 2 * (*d.DelaysMatrix)[requesterRegionTag][requesterRegionTag]
		if requesterRegionTag != myRegionTag {
			delay += latencies[requesterClosestNode][d.MyClosestNode] + latencies[d.MyClosestNode][requesterClosestNode]
		}

		log.Infof("adding %f ms from client to node", delay)
	} else {
		// requests between nodes are already being delay at the OS level
		delay = 0
	}

	return
}

func (d *S2DelayedCommsManager) getWSDelay(requesterCell s2.CellID,
	requesterClosestNode int) (myRegionTag, requesterRegionTag string, delay float64) {
	d.RLock()
	myRegionTag = TranslateCellToRegion(d.GetCellID())
	d.RUnlock()
	requesterRegionTag = TranslateCellToRegion(requesterCell)

	// we will apply the delay from client -> closestNode and if closest node is not the one being used
	// add the delay from closestNode -> targetNode

	delay = 2 * (*d.DelaysMatrix)[requesterRegionTag][requesterRegionTag]

	if requesterRegionTag != myRegionTag {
		delay += latencies[requesterClosestNode][d.MyClosestNode] + latencies[d.MyClosestNode][requesterClosestNode]
	}

	log.Infof("adding %f ms from client to node", delay)

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

func (d *S2DelayedCommsManager) GetCellID() s2.CellID {
	d.RLock()
	cellID := d.CellId
	d.RUnlock()

	return cellID
}

func (d *S2DelayedCommsManager) SetCellID(cellID s2.CellID) {
	d.Lock()
	d.CellId = cellID
	d.Unlock()
}

const (
	defaultNodeLatenciesPath = "/service/lats.txt"
)

func loadNodeLatencies() (latencies map[int][]float64) {
	latencies = map[int][]float64{}

	var (
		nodeLatenciesPath string
		ok                bool
	)

	if nodeLatenciesPath, ok = os.LookupEnv("LAT"); !ok {
		nodeLatenciesPath = defaultNodeLatenciesPath
	}

	f, err := os.Open(nodeLatenciesPath)
	if err != nil {
		log.Panic(err)
	}

	defer f.Close()

	scanner := bufio.NewScanner(f)
	scanner.Split(bufio.ScanLines)

	number := 0
	for scanner.Scan() {
		line := scanner.Text()
		line = strings.TrimSpace(line)
		lats := strings.Split(line, " ")

		nodeLats := make([]float64, len(lats))
		for i, lat := range lats {
			nodeLats[i], err = strconv.ParseFloat(lat, 64)
			if err != nil {
				log.Panic(err)
			}
		}

		latencies[number] = nodeLats

		number++
	}

	return
}

const (
	defaultNodeLocationsPath = "/service/locations.json"
)

func getClosestNode(cellID s2.CellID) int {
	var (
		nodeLocationsPath string
		ok                bool
	)

	if nodeLocationsPath, ok = os.LookupEnv("LOCATIONS"); !ok {
		nodeLocationsPath = defaultNodeLocationsPath
	}

	f, err := os.Open(nodeLocationsPath)
	if err != nil {
		log.Panic(err)
	}
	defer f.Close()

	var locations map[string]struct {
		Lat float64
		Lng float64
	}

	var data map[string]interface{}

	err = json.NewDecoder(f).Decode(&data)
	if err != nil {
		log.Panic(err)
	}

	err = mapstructure.Decode(data, &locations)
	if err != nil {
		log.Panic(err)
	}

	myCell := s2.CellFromCellID(cellID)
	minDist := -1.
	closestNode := -1

	for nodeNumString, latLng := range locations {
		nodeCell := s2.CellFromLatLng(s2.LatLngFromDegrees(latLng.Lat, latLng.Lng))
		dist := chordAngleToKM(myCell.DistanceToCell(nodeCell))

		if minDist == -1 || dist < minDist {
			minDist = dist
			closestNode, err = strconv.Atoi(nodeNumString)
			if err != nil {
				log.Panic(err)
			}
		}
	}

	return closestNode
}

const (
	earthRadius = 6_378
)

func chordAngleToKM(angle s1.ChordAngle) float64 {
	return angle.Angle().Radians() * earthRadius
}
