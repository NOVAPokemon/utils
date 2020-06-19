package websockets

import (
	"encoding/json"
	"strings"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	log "github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Lobby maintains the connections from both trainers and the status of the battle
type Lobby struct {
	Id primitive.ObjectID

	TrainersJoined     int
	TrainerUsernames   []string
	TrainerInChannels  []chan *string
	TrainerOutChannels []chan GenericMsg

	trainerConnections    []*websocket.Conn
	EndConnectionChannels []chan struct{}
	changeLobbyLock       *sync.Mutex
	startOnce             sync.Once
	finishOnce            sync.Once
	closeOnce             sync.Once
	Started               chan struct{}
	Finished              chan struct{}
	capacity              int
}

func NewLobby(id primitive.ObjectID, capacity int) *Lobby {
	return &Lobby{
		capacity:              capacity,
		Id:                    id,
		TrainersJoined:        0,
		TrainerUsernames:      make([]string, capacity),
		trainerConnections:    make([]*websocket.Conn, capacity),
		TrainerInChannels:     make([]chan *string, capacity),
		TrainerOutChannels:    make([]chan GenericMsg, capacity),
		EndConnectionChannels: make([]chan struct{}, capacity),
		finishOnce:            sync.Once{},
		startOnce:             sync.Once{},
		closeOnce:             sync.Once{},
		Started:               make(chan struct{}),
		Finished:              make(chan struct{}),
		changeLobbyLock:       &sync.Mutex{},
	}
}

func AddTrainer(lobby *Lobby, username string, trainerConn *websocket.Conn) (int, error) {
	lobby.changeLobbyLock.Lock()
	defer lobby.changeLobbyLock.Unlock()

	if lobby.TrainersJoined >= lobby.capacity {
		return -1, NewLobbyIsFullError(lobby.Id.Hex())
	}

	select {
	case <-lobby.Started:
		return - 1, NewLobbyStartedError(lobby.Id.Hex())
	default:
		trainerChanIn := make(chan *string)
		trainerChanOut := make(chan GenericMsg)
		lobby.TrainerUsernames[lobby.TrainersJoined] = username
		lobby.TrainerInChannels[lobby.TrainersJoined] = trainerChanIn
		lobby.TrainerOutChannels[lobby.TrainersJoined] = trainerChanOut
		lobby.trainerConnections[lobby.TrainersJoined] = trainerConn
		lobby.EndConnectionChannels[lobby.TrainersJoined] = make(chan struct{})
		go HandleReceiveLobby(trainerConn, trainerChanIn, lobby.EndConnectionChannels[lobby.TrainersJoined], lobby.Finished)
		go HandleSendLobby(trainerConn, trainerChanOut, lobby.EndConnectionChannels[lobby.TrainersJoined], lobby.Finished)
		lobby.TrainersJoined++
		return lobby.TrainersJoined, nil
	}
}

func HandleReceiveLobby(conn *websocket.Conn, outChannel chan *string, endConnection chan struct{},
	finished chan struct{}) {
	err := HandleRecv(conn, outChannel, endConnection)

	if err != nil {
		select {
		case <-finished:
			log.Info("lobby read routine finished properly")
		default:
			log.Warn(err)
			log.Warn("closed lobby because could not read")
		}
	}
}

// server side
func HandleSendLobby(conn *websocket.Conn, inChannel chan GenericMsg, endConnection chan struct{}, finished chan struct{}) {
	err := HandleSend(conn, inChannel, endConnection)

	if err != nil {
		select {
		case <-finished:
			log.Info("lobby write routine finished properly")
		default:
			log.Warn(err)
			log.Warn("closed lobby because could not write")
		}
	}
}

func StartLobby(lobby *Lobby) {
	lobby.startOnce.Do(func() {
		close(lobby.Started)
	})
}

func FinishLobby(lobby *Lobby) {
	lobby.finishOnce.Do(func() {
		close(lobby.Finished)
	})
}

func CloseLobbyConnections(lobby *Lobby) {
	lobby.closeOnce.Do(func() {
		lobby.changeLobbyLock.Lock()
		defer lobby.changeLobbyLock.Unlock()
		for i := 0; i < len(lobby.EndConnectionChannels); i++ {
			closeConnectionThroughChannel(lobby.trainerConnections[i], lobby.EndConnectionChannels[i])
		}
	})
}

func GetTrainersJoined(lobby *Lobby) int {
	lobby.changeLobbyLock.Lock()
	defer lobby.changeLobbyLock.Unlock()
	return lobby.TrainersJoined
}

const (
	PongWait   = 2 * time.Second
	PingPeriod = (PongWait * 9) / 10
)

func ParseMessage(msg *string) (*Message, error) {
	if msg == nil {
		return nil, wrapMsgParsingError(ErrorMessageNil)
	}
	toReturn := &Message{}
	if err := json.Unmarshal([]byte(*msg), toReturn); err != nil {
		return nil, wrapMsgParsingError(ErrorMessageNil)
	}
	return toReturn, nil
}

func HandleSend(conn *websocket.Conn, outChannel chan GenericMsg, endConnection chan struct{}) error {
	pingTicker := time.NewTicker(PingPeriod)
	conn.SetPongHandler(func(_ string) error {
		// log.Info("Received pong")
		return conn.SetReadDeadline(time.Now().Add(PongWait))
	})
	for {
		select {
		case <-pingTicker.C:
			// log.Warn("Pinging")
			if err := conn.WriteMessage(websocket.PingMessage, []byte{}); err != nil {
				closeConnectionThroughChannel(conn, endConnection)
				return wrapHandleSendError(err)
			}
		case msg := <-outChannel:
			err := conn.WriteMessage(msg.MsgType, msg.Data)
			if err != nil {
				closeConnectionThroughChannel(conn, endConnection)
				return wrapHandleSendError(err)
			}
		case <-endConnection:
			log.Info("Send routine finishing properly")
			return nil
		}

	}
}

func HandleRecv(conn *websocket.Conn, inChannel chan *string, endConnection chan struct{}) error {
	defer close(inChannel)
	for {
		select {
		case <-endConnection:
			return nil
		default:
			_, message, err := conn.ReadMessage()
			if err != nil {
				select {
				case <-endConnection:
					log.Info("Receive routine finishing properly")
					return nil
				default:
					log.Warn("Closing finish channel and connection")
					closeConnectionThroughChannel(conn, endConnection)
					return wrapHandleReceiveError(err)
				}
			} else {
				msg := strings.TrimSpace(string(message))
				inChannel <- &msg
			}
		}
	}
}

func closeConnectionThroughChannel(conn *websocket.Conn, endConnection chan struct{}) {
	close(endConnection)
	closeConnection(conn)
}

func closeConnection(conn *websocket.Conn) {
	if conn == nil {
		return
	}
	if err := conn.Close(); err != nil {
		log.Error(err)
	}
}
