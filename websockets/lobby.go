package websockets

import (
	"github.com/gorilla/websocket"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Lobby maintains the connections from both trainers and the status of the battle
type Lobby struct {
	Id primitive.ObjectID

	TrainersJoined     int
	TrainerUsernames   [2]string
	TrainerInChannels  [2]*chan *string
	TrainerOutChannels [2]*chan *string

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
		TrainerOutChannels:    [2]*chan *string{},
		EndConnectionChannels: [2]chan struct{}{make(chan struct{}), make(chan struct{})},

		Started:  false,
		Finished: false,
	}
}

func AddTrainer(lobby *Lobby, username string, trainerConn *websocket.Conn) {

	trainerChanIn := make(chan *string)
	trainerChanOut := make(chan *string)

	go handleRecv(trainerConn, trainerChanIn, lobby.EndConnectionChannels[lobby.TrainersJoined], &lobby.Finished)
	go handleSend(trainerConn, trainerChanOut, lobby.EndConnectionChannels[lobby.TrainersJoined], &lobby.Finished)

	lobby.TrainerUsernames[lobby.TrainersJoined] = username
	lobby.TrainerInChannels[lobby.TrainersJoined] = &trainerChanIn
	lobby.TrainerOutChannels[lobby.TrainersJoined] = &trainerChanOut
	lobby.trainerConnections[lobby.TrainersJoined] = trainerConn
	lobby.TrainersJoined++
}

func CloseLobby(lobby *Lobby) {
	lobby.Finished = true
	closeConnectionThroughChannel(lobby.trainerConnections[0], lobby.EndConnectionChannels[0])
	closeConnectionThroughChannel(lobby.trainerConnections[1], lobby.EndConnectionChannels[1])
}
