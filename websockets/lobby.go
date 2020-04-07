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
		Started:               false,
		Finished:              false,
	}
}

func AddTrainer(lobby *Lobby, username string, trainerConn *websocket.Conn) {

	trainerChanIn := make(chan *string)
	trainerChanOut := make(chan *string)

	go handleRecv(trainerConn, trainerChanIn, lobby.EndConnectionChannels[lobby.TrainersJoined])
	go handleSend(trainerConn, trainerChanOut, lobby.EndConnectionChannels[lobby.TrainersJoined])

	lobby.TrainerUsernames[lobby.TrainersJoined] = username
	lobby.TrainerInChannels[lobby.TrainersJoined] = &trainerChanIn
	lobby.TrainerOutChannels[lobby.TrainersJoined] = &trainerChanOut
	lobby.trainerConnections[lobby.TrainersJoined] = trainerConn
	lobby.TrainersJoined++
}

func CloseLobby(lobby *Lobby) {
	closeConnection(lobby.trainerConnections[0], lobby.EndConnectionChannels[0])
	closeConnection(lobby.trainerConnections[1], lobby.EndConnectionChannels[1])
}

func closeConnection(conn *websocket.Conn, endConnection chan struct{}) {
	endChannel(endConnection)
	if conn != nil {
		conn.Close()
	}
}

func handleSend(conn *websocket.Conn, inChannel chan *string, endConnection chan struct{}) {
	defer close(inChannel)
	defer log.Warn("Closing send routine")

	for {
		select {
		case msg := <-inChannel:
			err := conn.WriteMessage(websocket.TextMessage, []byte(*msg))
			if err != nil {
				log.Error(err)
				log.Error("closed lobby because could not write")
				closeConnection(conn, endConnection)
				return
			} else {
				log.Infof("Wrote %s into the channel", *msg)
			}
		case <-endConnection:
			return
		}

	}
}

func handleRecv(conn *websocket.Conn, outChannel chan *string, endConnection chan struct{}) {
	defer close(outChannel)
	defer log.Warn("Closing recv routine")

	for {
		select {
		case <-endConnection:
			return
		default:
			_, message, err := conn.ReadMessage()

			if err != nil {
				log.Error("closed lobby because could not read")
				closeConnection(conn, endConnection)
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

func endChannel(channel chan struct{}) {
	select {
	case <-channel:
		return
	default:
		close(channel)
	}
}
