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

	Started  bool
	Finished bool
}

func NewLobby(id primitive.ObjectID, ) *Lobby {

	return &Lobby{
		Id:                 id,
		trainerConnections: make([]*websocket.Conn, 0),
		TrainerInChannels:  make([]*chan *string, 0),
		TrainerOutChannels: make([]*chan *string, 0),
		Started:            false,
		Finished:           false,
	}
}

func AddTrainer(lobby *Lobby, trainer utils.Trainer, trainerConn *websocket.Conn) {

	trainerChanIn := make(chan *string)
	trainerChanOut := make(chan *string)

	go handleRecv(trainerConn, trainerChanIn)
	go handleSend(trainerConn, trainerChanOut)

	lobby.Trainers = append(lobby.Trainers, &trainer)
	lobby.TrainerInChannels = append(lobby.TrainerInChannels, &trainerChanIn)
	lobby.TrainerOutChannels = append(lobby.TrainerOutChannels, &trainerChanOut)

}

func CloseLobby(lobby *Lobby) {

	// TODO
}

func handleSend(conn *websocket.Conn, channel chan *string) {

	for {
		msg := <-channel
		err := conn.WriteMessage(websocket.TextMessage, []byte(*msg))
		if err != nil {
			log.Error("write err:", err)
		} else {
			log.Debugf("Wrote %s into the channel", msg)
		}

	}

}

func handleRecv(conn *websocket.Conn, channel chan *string) {

	for {
		_, message, err := conn.ReadMessage()

		if err != nil {
			log.Println(err)
		} else {
			msg := strings.TrimSpace(string(message))
			log.Infof("Message received: %s", msg)
			channel <- &msg
		}
	}
}