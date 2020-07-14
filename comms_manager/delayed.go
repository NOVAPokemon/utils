package comms_manager

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/golang/geo/s2"
)

type LocationLatency struct {
	locationName       string
	directLatency      int
	subLocationLatency int
	subLocations       []LocationLatency
}

type DelayedCommsManager struct {
	locations           []LocationLatency
	myLocation          string
	positionInitialized bool
}

func (w *DelayedCommsManager) setLocation(location s2.LatLng) {
	w.myLocation = location
	if !w.positionInitialized {
		w.positionInitialized = true
	}
}

// For now we only take into account two levels (Datacenter and directly connected edge nodes)
func (w *DelayedCommsManager) getDelay() time.Duration {
	found := false
	delayToApply := 0

	closest := -1
	delayToApply := 0

	for _, location := range w.locations {
		location.
		for _, subLocation := range location.subLocations {
			if subLocation.directLatency + location.directLatency < currClosest {

			}
		}
	}

	for serverName, delay := range w.delayConfigs.delays {
		if strings.Contains(serverNameToFind, serverName) {
			found = true
			delayToApply = delay
		}
	}
	if !found {
		if defaultDelay, ok := w.delayConfigs.delays[defaultDelayField]; ok {
			if !ok {
				panic(fmt.Sprintf("could not find delay for %s: %+v", serverNameToFind, w.delayConfigs.delays))
			}
			delayToApply = defaultDelay
		}
	}

	return time.Duration(delayToApply) * time.Millisecond
}

func (w *DelayedCommsManager) WriteMessageToConn(conn *WsConnWithServerName, msgType int, data []byte) error {
	delay := w.getDelayFromServerName(conn.ServerName)

	time.Sleep(delay * time.Millisecond)
	return conn.WriteMessage(msgType, data)
}

func (w *DelayedCommsManager) DoHTTPRequest(client *http.Client, req *http.Request) (*http.Response, error) {
	delay := w.getDelayFromServerName(req.URL.Host)
	time.Sleep(delay * time.Millisecond)
	return client.Do(req)
}
