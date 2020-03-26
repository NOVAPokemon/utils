package websockets

import (
	"github.com/NOVAPokemon/utils"
	"github.com/gorilla/websocket"
	log "github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"strings"
)

// Lobby maintains the connections from both trainers and the status of the battle
type Lobby struct {
	Id primitive.ObjectID

	Trainers           []*utils.Trainer
	TrainerInChannels  []*chan *string
	TrainerOutChannels []*chan *string

	trainerConnections []*websocket.Conn

	EndConnectionChannel chan struct{}

	Started  bool
	Finished bool
}

func NewLobby(id primitive.ObjectID, ) *Lobby {
	return &Lobby{
		Id:                   id,
		trainerConnections:   make([]*websocket.Conn, 0),
		TrainerInChannels:    make([]*chan *string, 0),
		TrainerOutChannels:   make([]*chan *string, 0),
		EndConnectionChannel: make(chan struct{}),
		Started:              false,
		Finished:             false,
	}
}

func AddTrainer(lobby *Lobby, trainer utils.Trainer, trainerConn *websocket.Conn) {

	trainerChanIn := make(chan *string)
	trainerChanOut := make(chan *string)

	go handleRecv(trainerConn, trainerChanIn, lobby.EndConnectionChannel)
	go handleSend(trainerConn, trainerChanOut, lobby.EndConnectionChannel)

	lobby.Trainers = append(lobby.Trainers, &trainer)
	lobby.TrainerInChannels = append(lobby.TrainerInChannels, &trainerChanIn)
	lobby.TrainerOutChannels = append(lobby.TrainerOutChannels, &trainerChanOut)

	lobby.trainerConnections = append(lobby.trainerConnections, trainerConn)
}

func CloseLobby(lobby *Lobby) {
	log.Warn("Triggering end connection on remaining go routines...")
	close(lobby.EndConnectionChannel)

	lobby.trainerConnections[0].Close()
	lobby.trainerConnections[1].Close()
}

func handleSend(conn *websocket.Conn, inChannel chan *string, endConnection chan struct{}) {
	defer close(inChannel)
	defer log.Warn("Closing send routine")

	for {
		select {
		case msg := <-inChannel:
			err := conn.WriteMessage(websocket.TextMessage, []byte(*msg))
			if err != nil {
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
