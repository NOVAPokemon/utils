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

	changeLobbyLock *sync.Mutex
	TrainersJoined  int

	closeChannelsOnce     []*sync.Once
	EndConnectionChannels []chan struct{}

	TrainerUsernames   []string
	TrainerInChannels  []chan *string
	TrainerOutChannels []chan GenericMsg
	trainerConnections []*websocket.Conn

	startOnce *sync.Once
	Started   chan struct{}

	finishOnce *sync.Once
	Finished   chan struct{}

	capacity int
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
		closeChannelsOnce:     make([]*sync.Once, capacity),
		finishOnce:            &sync.Once{},
		startOnce:             &sync.Once{},
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
		trainerNum := lobby.TrainersJoined
		trainerChanIn := make(chan *string)
		trainerChanOut := make(chan GenericMsg)
		lobby.TrainerUsernames[trainerNum] = username
		lobby.TrainerInChannels[trainerNum] = trainerChanIn
		lobby.TrainerOutChannels[trainerNum] = trainerChanOut
		lobby.trainerConnections[trainerNum] = trainerConn
		lobby.EndConnectionChannels[trainerNum] = make(chan struct{})
		lobby.closeChannelsOnce[trainerNum] = &sync.Once{}
		go HandleReceiveLobby(lobby, trainerNum)
		go HandleSendLobby(lobby, trainerNum)
		lobby.TrainersJoined++
		return lobby.TrainersJoined, nil
	}
}

func HandleReceiveLobby(lobby *Lobby, i int) {
	err := HandleRecv(lobby, i)

	if err != nil {
		select {
		case <-lobby.Finished:
			log.Info("lobby read routine finished properly")
		default:
			log.Warn(err)
			log.Warn("closed lobby because could not read")
		}
	}
}

// server side
func HandleSendLobby(lobby *Lobby, i int) {
	err := HandleSend(lobby, i)

	if err != nil {
		select {
		case <-lobby.EndConnectionChannels[i]:
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
	lobby.changeLobbyLock.Lock()
	defer lobby.changeLobbyLock.Unlock()
	for i := 0; i < len(lobby.EndConnectionChannels); i++ {
		closeConnectionThroughChannel(lobby.closeChannelsOnce[i], lobby.trainerConnections[i], lobby.EndConnectionChannels[i])
	}
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

func HandleSend(lobby *Lobby, trainerNum int) error {
	pingTicker := time.NewTicker(PingPeriod)

	conn := lobby.trainerConnections[trainerNum]
	endConnection := lobby.EndConnectionChannels[trainerNum]
	outChannel := lobby.TrainerOutChannels[trainerNum]
	closeOnce := lobby.closeChannelsOnce[trainerNum]

	conn.SetPongHandler(func(_ string) error {
		// log.Info("Received pong")
		return conn.SetReadDeadline(time.Now().Add(PongWait))
	})
	for {
		select {
		case <-pingTicker.C:
			// log.Warn("Pinging")
			if err := conn.WriteMessage(websocket.PingMessage, []byte{}); err != nil {
				closeConnectionThroughChannel(closeOnce, conn, endConnection)
				return wrapHandleSendError(err)
			}
		case msg := <-outChannel:
			err := conn.WriteMessage(msg.MsgType, msg.Data)
			if err != nil {
				closeConnectionThroughChannel(closeOnce, conn, endConnection)
				return wrapHandleSendError(err)
			}
		case <-endConnection:
			log.Info("Send routine finishing properly")
			return nil
		}

	}
}

func HandleRecv(lobby *Lobby, trainerNum int) error {

	conn := lobby.trainerConnections[trainerNum]
	endConnection := lobby.EndConnectionChannels[trainerNum]
	inChannel := lobby.TrainerInChannels[trainerNum]
	closeOnce := lobby.closeChannelsOnce[trainerNum]

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
					closeConnectionThroughChannel(closeOnce, conn, endConnection)
					return wrapHandleReceiveError(err)
				}
			} else {
				msg := strings.TrimSpace(string(message))
				inChannel <- &msg
			}
		}
	}
}

func closeConnectionThroughChannel(closeOnce *sync.Once, conn *websocket.Conn, endConnection chan struct{}) {
	closeOnce.Do(func() {
		endChannel(endConnection)
		closeConnection(conn)
	})
}

func endChannel(channel chan struct{}) {
	if channel == nil {
		return
	}

	select {
	case <-channel:
		return
	default:
		close(channel)
	}
}

func closeConnection(conn *websocket.Conn) {
	if conn == nil {
		return
	}
	if err := conn.Close(); err != nil {
		log.Error(err)
	}
}
