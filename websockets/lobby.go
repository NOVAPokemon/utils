package websockets

import (
	"sync"

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
	joinLock              *sync.Mutex
	startOnce             sync.Once
	finishOnce            sync.Once
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
		Started:               make(chan struct{}),
		Finished:              make(chan struct{}),
		joinLock:              &sync.Mutex{},
	}
}

func AddTrainer(lobby *Lobby, username string, trainerConn *websocket.Conn) (int, error) {
	lobby.joinLock.Lock()
	defer lobby.joinLock.Unlock()

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
	for i := 0; i < len(lobby.EndConnectionChannels); i++ {
		closeConnectionThroughChannel(lobby.trainerConnections[i], lobby.EndConnectionChannels[i])
	}
}
