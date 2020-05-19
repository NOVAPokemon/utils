package websockets

import (
	"github.com/gorilla/websocket"
	log "github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Lobby maintains the connections from both trainers and the status of the battle
type Lobby struct {
	Id primitive.ObjectID

	TrainersJoined     int
	TrainerUsernames   [2]string
	TrainerInChannels  [2]*chan *string
	TrainerOutChannels [2]*chan GenericMsg

	trainerConnections    [2]*websocket.Conn
	EndConnectionChannels [2]chan struct{}

	Started  bool
	Finished bool
}

func NewLobby(id primitive.ObjectID) *Lobby {
	return &Lobby{
		Id:                    id,
		TrainersJoined:        0,
		TrainerUsernames:      [2]string{},
		trainerConnections:    [2]*websocket.Conn{},
		TrainerInChannels:     [2]*chan *string{},
		TrainerOutChannels:    [2]*chan GenericMsg{},
		EndConnectionChannels: [2]chan struct{}{make(chan struct{}), make(chan struct{})},

		Started:  false,
		Finished: false,
	}
}

func AddTrainer(lobby *Lobby, username string, trainerConn *websocket.Conn) {

	trainerChanIn := make(chan *string)
	trainerChanOut := make(chan GenericMsg)

	go HandleReceiveLobby(trainerConn, trainerChanIn, lobby.EndConnectionChannels[lobby.TrainersJoined], &lobby.Finished)
	go HandleSendLobby(trainerConn, trainerChanOut, lobby.EndConnectionChannels[lobby.TrainersJoined], &lobby.Finished)

	lobby.TrainerUsernames[lobby.TrainersJoined] = username
	lobby.TrainerInChannels[lobby.TrainersJoined] = &trainerChanIn
	lobby.TrainerOutChannels[lobby.TrainersJoined] = &trainerChanOut
	lobby.trainerConnections[lobby.TrainersJoined] = trainerConn
	lobby.TrainersJoined++
}

func HandleReceiveLobby(conn *websocket.Conn, outChannel chan *string, endConnection chan struct{}, finished *bool) {
	err := HandleRecv(conn, outChannel, endConnection)

	if err != nil {
		if *finished {
			log.Info("lobby read routine finished properly")
		} else {
			log.Warn(err)
			log.Warn("closed lobby because could not read")
		}
	}
}

// server side
func HandleSendLobby(conn *websocket.Conn, inChannel chan GenericMsg, endConnection chan struct{}, finished *bool) {
	err := HandleSend(conn, inChannel, endConnection)

	if err != nil {
		if *finished {
			log.Info("lobby write routine finished properly")
		} else {
			log.Warn(err)
			log.Warn("closed lobby because could not write")
		}
	}
}

func CloseLobby(lobby *Lobby) {
	lobby.Finished = true
	closeConnectionThroughChannel(lobby.trainerConnections[0], lobby.EndConnectionChannels[0])
	closeConnectionThroughChannel(lobby.trainerConnections[1], lobby.EndConnectionChannels[1])
}
