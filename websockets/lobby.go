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

const (
	PongWait   = 2 * time.Second
	PingPeriod = (PongWait * 9) / 10
)

// Lobby maintains the connections from both trainers and the status of the battle
type Lobby struct {
	Id              primitive.ObjectID
	changeLobbyLock *sync.Mutex
	TrainersJoined  int
	Started         chan struct{}
	Finished        chan struct{}
	Capacity        int

	TrainerUsernames      []string
	TrainerInChannels     []chan string
	DoneListeningFromConn []chan interface{}
	DoneWritingToConn     []chan interface{}
	TrainerOutChannels    []chan GenericMsg
	trainerConnections    []*websocket.Conn
	finishOnce            sync.Once
}

func NewLobby(id primitive.ObjectID, capacity int) *Lobby {
	return &Lobby{
		Capacity:              capacity,
		Id:                    id,
		TrainersJoined:        0,
		TrainerUsernames:      make([]string, capacity),
		trainerConnections:    make([]*websocket.Conn, capacity),
		TrainerInChannels:     make([]chan string, capacity),
		DoneListeningFromConn: make([]chan interface{}, capacity),
		TrainerOutChannels:    make([]chan GenericMsg, capacity),
		DoneWritingToConn:     make([]chan interface{}, capacity),
		Started:               make(chan struct{}),
		Finished:              make(chan struct{}),
		changeLobbyLock:       &sync.Mutex{},
		finishOnce:            sync.Once{},
	}
}

func AddTrainer(lobby *Lobby, username string, trainerConn *websocket.Conn,
	commsManager CommunicationManager) (int, error) {
	lobby.changeLobbyLock.Lock()
	defer lobby.changeLobbyLock.Unlock()

	if lobby.TrainersJoined >= lobby.Capacity {
		return -1, NewLobbyIsFullError(lobby.Id.Hex())
	}

	select {
	case <-lobby.Started:
		return - 1, NewLobbyStartedError(lobby.Id.Hex())
	case <-lobby.Finished:
		return - 1, NewLobbyFinishedError(lobby.Id.Hex())
	default:
		trainerNum := lobby.TrainersJoined
		trainerChanIn := make(chan string)
		trainerChanOut := make(chan GenericMsg)

		lobby.TrainerUsernames[trainerNum] = username
		lobby.TrainerInChannels[trainerNum] = trainerChanIn
		lobby.TrainerOutChannels[trainerNum] = trainerChanOut
		lobby.trainerConnections[trainerNum] = trainerConn
		lobby.DoneListeningFromConn[trainerNum] = RecvFromConnToChann(lobby, trainerNum, commsManager)
		lobby.DoneWritingToConn[trainerNum] = sendFromChanToConn(lobby, trainerNum, commsManager)
		lobby.TrainersJoined++
		return lobby.TrainersJoined, nil
	}
}

func sendFromChanToConn(lobby *Lobby, trainerNum int, writer CommunicationManager) (done chan interface{}) {
	done = make(chan interface{})
	go func() {
		pingTicker := time.NewTicker(PingPeriod)
		conn := lobby.trainerConnections[trainerNum]
		outChannel := lobby.TrainerOutChannels[trainerNum]
		conn.SetPongHandler(func(_ string) error {
			return conn.SetReadDeadline(time.Now().Add(PongWait))
		})

		defer close(done)

		for {
			select {
			case <-pingTicker.C:
				err := writer.WriteGenericMessageToConn(conn, GenericMsg{
					MsgType: websocket.PingMessage,
					Data:    nil,
				})
				if err != nil {
					log.Warn(err)
					return
				}
			case msg, ok := <-outChannel:
				if !ok {
					continue
				}
				err := writer.WriteGenericMessageToConn(conn, msg)
				if err != nil {
					log.Warn(err)
					return
				}
			case <-lobby.Finished:
				log.Info("Send routine finishing")
				return
			}
		}
	}()
	return done
}

func RecvFromConnToChann(lobby *Lobby, trainerNum int,
	manager CommunicationManager) (done chan interface{}) {
	done = make(chan interface{})
	go func() {
		conn := lobby.trainerConnections[trainerNum]
		inChannel := lobby.TrainerInChannels[trainerNum]
		defer close(done)
		for {
			message, err := manager.ReadTextMessageFromConn(conn)
			if err != nil {
				log.Info("Receive routine finishing because connection was closed")
				return
			}
			msg := strings.TrimSpace(string(message))
			select {
			case <-lobby.Finished:
				log.Infof("Could not send message %s because finish channel ended meanwhile", msg)
				return
			case inChannel <- msg:
				log.Debugf("Received message from Websockets")
			}
		}
	}()
	return done
}

func StartLobby(lobby *Lobby) {
	close(lobby.Started)
}

func FinishLobby(lobby *Lobby) {
	lobby.finishOnce.Do(func() {
		lobby.changeLobbyLock.Lock()
		defer lobby.changeLobbyLock.Unlock()
		close(lobby.Finished)
		for i := 0; i < lobby.TrainersJoined; i++ {
			if err := lobby.trainerConnections[i].Close(); err != nil {
				log.Error(err)
			}
			<-lobby.DoneWritingToConn[i]
			<-lobby.DoneListeningFromConn[i]
			close(lobby.TrainerOutChannels[i])
			close(lobby.TrainerInChannels[i])
		}
	})
}

func GetTrainersJoined(lobby *Lobby) int {
	lobby.changeLobbyLock.Lock()
	defer lobby.changeLobbyLock.Unlock()
	return lobby.TrainersJoined
}

func ParseMessage(msg string) (*Message, error) {
	toReturn := &Message{}
	if err := json.Unmarshal([]byte(msg), toReturn); err != nil {
		log.Error(msg)
		return nil, wrapMsgParsingError(err)
	}
	return toReturn, nil
}
