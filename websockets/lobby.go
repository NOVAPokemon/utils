package websockets

import (
	"github.com/gorilla/websocket"
	log "github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Lobby maintains the connections from both trainers and the status of the battle
type Lobby struct {
	Id primitive.ObjectID

	TrainersJoined     int64
	TrainerUsernames   [2]string
	TrainerInChannels  [2]*chan *string
	TrainerOutChannels [2]*SyncChannel

	trainerConnections    [2]*websocket.Conn
	EndConnectionChannels [2]chan struct{}

	Started  bool
	Finished chan struct{}
}

func NewLobby(id primitive.ObjectID) *Lobby {
	return &Lobby{
		Id:                    id,
		TrainersJoined:        0,
		TrainerUsernames:      [2]string{},
		trainerConnections:    [2]*websocket.Conn{},
		TrainerInChannels:     [2]*chan *string{},
		TrainerOutChannels:    [2]*SyncChannel{},
		EndConnectionChannels: [2]chan struct{}{make(chan struct{}), make(chan struct{})},

		Started:  false,
		Finished: make(chan struct{}),
	}
}

func AddTrainer(lobby *Lobby, username string, trainerConn *websocket.Conn) int64 {
	trainerChanIn := make(chan *string)
	trainerChanOut := NewSyncChannel(make(chan GenericMsg))

	go HandleReceiveLobby(trainerConn, trainerChanIn, lobby.EndConnectionChannels[lobby.TrainersJoined], lobby.Finished)
	go HandleSendLobby(trainerConn, trainerChanOut, lobby.EndConnectionChannels[lobby.TrainersJoined], lobby.Finished)

	lobby.TrainerUsernames[lobby.TrainersJoined] = username
	lobby.TrainerInChannels[lobby.TrainersJoined] = &trainerChanIn
	lobby.TrainerOutChannels[lobby.TrainersJoined] = trainerChanOut
	lobby.trainerConnections[lobby.TrainersJoined] = trainerConn
	lobby.TrainersJoined++
	return lobby.TrainersJoined
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
func HandleSendLobby(conn *websocket.Conn, inChannel *SyncChannel, endConnection chan struct{}, finished chan struct{}) {
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

func CloseLobby(lobby *Lobby) {
	closeConnectionThroughChannel(lobby.trainerConnections[0], lobby.EndConnectionChannels[0])
	closeConnectionThroughChannel(lobby.trainerConnections[1], lobby.EndConnectionChannels[1])
}
