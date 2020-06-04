package websockets

import (
	"github.com/gorilla/websocket"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type RaidLobby struct {
	Id primitive.ObjectID

	TrainersJoined     int
	TrainerUsernames   []string
	TrainerInChannels  []*chan *string
	TrainerOutChannels []*SyncChannel

	trainerConnections    []*websocket.Conn
	EndConnectionChannels []chan struct{}
	ActiveConnections     int
	Started               bool
	Finished              bool
}

func NewRaidLobby(id primitive.ObjectID, expectedCapacity int) *RaidLobby {
	return &RaidLobby{
		Id:                    id,
		TrainersJoined:        0,
		TrainerUsernames:      make([]string, 0, expectedCapacity),
		trainerConnections:    make([]*websocket.Conn, 0, expectedCapacity),
		TrainerInChannels:     make([]*chan *string, 0, expectedCapacity),
		TrainerOutChannels:    make([]*SyncChannel, 0, expectedCapacity),
		EndConnectionChannels: make([]chan struct{}, 0, expectedCapacity),
		ActiveConnections:     0,
		Started:               false,
		Finished:              false,
	}
}

func (lobby *RaidLobby) AddTrainer(username string, trainerConn *websocket.Conn) {

	trainerChanIn := make(chan *string)
	trainerChanOut := NewSyncChannel(make(chan GenericMsg))
	endChan := make(chan struct{})

	lobby.TrainerUsernames = append(lobby.TrainerUsernames, username)
	lobby.TrainerInChannels = append(lobby.TrainerInChannels, &trainerChanIn)
	lobby.TrainerOutChannels = append(lobby.TrainerOutChannels, trainerChanOut)
	lobby.trainerConnections = append(lobby.trainerConnections, trainerConn)
	lobby.EndConnectionChannels = append(lobby.EndConnectionChannels, endChan)

	go HandleReceiveLobby(trainerConn, trainerChanIn, lobby.EndConnectionChannels[lobby.TrainersJoined], &lobby.Finished)
	go HandleSendLobby(trainerConn, trainerChanOut, lobby.EndConnectionChannels[lobby.TrainersJoined], &lobby.Finished)

	lobby.TrainersJoined++
	lobby.ActiveConnections++
}

func (lobby *RaidLobby) Close() {
	lobby.Finished = true
	for i := 0; i < len(lobby.EndConnectionChannels); i++ {
		closeConnectionThroughChannel(lobby.trainerConnections[i], lobby.EndConnectionChannels[i])
	}
}
