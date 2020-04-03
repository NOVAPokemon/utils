package websockets

import (
	"github.com/gorilla/websocket"
	log "github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"strings"
)

// Lobby maintains the connections from both trainers and the status of the battle
type Lobby struct {
	Id primitive.ObjectID

	TrainersJoined     int
	TrainerUsernames   [2]string
	TrainerInChannels  [2]*chan *string
	TrainerOutChannels [2]*chan *string

	trainerConnections [2]*websocket.Conn

	EndConnectionChannel chan struct{}

	Started  bool
	Finished bool
}

func NewLobby(id primitive.ObjectID) *Lobby {
	return &Lobby{
		Id:                   id,
		TrainersJoined:       0,
		TrainerUsernames:     [2]string{},
		trainerConnections:   [2]*websocket.Conn{},
		TrainerInChannels:    [2]*chan *string{},
		TrainerOutChannels:   [2]*chan *string{},
		EndConnectionChannel: make(chan struct{}),
		Started:              false,
		Finished:             false,
	}
}

func AddTrainer(lobby *Lobby, username string, trainerConn *websocket.Conn) {

	trainerChanIn := make(chan *string)
	trainerChanOut := make(chan *string)

	go handleRecv(trainerConn, trainerChanIn, lobby)
	go handleSend(trainerConn, trainerChanOut, lobby)

	lobby.TrainerUsernames[lobby.TrainersJoined] = username
	lobby.TrainerInChannels[lobby.TrainersJoined] = &trainerChanIn
	lobby.TrainerOutChannels[lobby.TrainersJoined] = &trainerChanOut
	lobby.trainerConnections[lobby.TrainersJoined] = trainerConn
	lobby.TrainersJoined++
}

func CloseLobby(lobby *Lobby) {
	endConnection(lobby)

	if lobby.trainerConnections[0] != nil {
		lobby.trainerConnections[0].Close()
	}

	if lobby.trainerConnections[1] != nil {
		lobby.trainerConnections[1].Close()
	}
}

func handleSend(conn *websocket.Conn, inChannel chan *string, lobby *Lobby) {
	defer close(inChannel)
	defer log.Warn("Closing send routine")

	for {
		select {
		case msg := <-inChannel:
			err := conn.WriteMessage(websocket.TextMessage, []byte(*msg))
			if err != nil {
				log.Error(err)
				log.Error("closed lobby because could not write")
				CloseLobby(lobby)
				return
			} else {
				log.Infof("Wrote %s into the channel", *msg)
			}
		case <-lobby.EndConnectionChannel:
			return
		}

	}
}

func handleRecv(conn *websocket.Conn, outChannel chan *string, lobby *Lobby) {
	defer close(outChannel)
	defer log.Warn("Closing recv routine")

	for {
		select {
		case <-lobby.EndConnectionChannel:
			return
		default:
			_, message, err := conn.ReadMessage()

			if err != nil {
				log.Error("closed lobby because could not read")
				CloseLobby(lobby)
				log.Error(err)
				return
			} else {
				msg := strings.TrimSpace(string(message))
				log.Infof("Message received: %s", msg)
				outChannel <- &msg
			}
		}
	}
}

func endConnection(lobby *Lobby) {
	select {
	case <-lobby.EndConnectionChannel:
		return
	default:
		close(lobby.EndConnectionChannel)
	}
}
