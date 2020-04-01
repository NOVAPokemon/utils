package clients

import (
	"bufio"
	"fmt"
	"github.com/NOVAPokemon/utils/websockets"
	"github.com/NOVAPokemon/utils/websockets/trades"
	"github.com/gorilla/websocket"
	log "github.com/sirupsen/logrus"
	"os"
)

func Send(conn *websocket.Conn, msg *string) {
	err := conn.WriteMessage(websocket.TextMessage, []byte(*msg))

	if err != nil {
		return
	} else {
		log.Debugf("Wrote %s into the channel", *msg)
	}
}

func ReadMessages(conn *websocket.Conn, finished chan struct{}) {
	defer close(finished)

	for {
		_, message, err := conn.ReadMessage()

		if err != nil {
			log.Error(err)
			return
		}

		msg := string(message)
		log.Debugf("Received %s from the websocket", msg)

		err, tradeMsg := websockets.ParseMessage(&msg)
		if err != nil {
			log.Error(err)
			continue
		}

		log.Infof("Message: %s", msg)

		if tradeMsg.MsgType == trades.FINISH {
			log.Info("Finished trade.")
			return
		}
	}
}

func WriteMessage(writeChannel chan *string) {
	for {
		reader := bufio.NewReader(os.Stdin)
		fmt.Print("Enter text: ")
		text, _ := reader.ReadString('\n')
		writeChannel <- &text
	}
}

func Finish() {
	log.Info("Finishing connection...")
}

func MainLoop(conn *websocket.Conn, writeChannel chan *string, finished chan struct{}) {
	for {
		select {
		case <-finished:
			Finish()
			return
		case msg := <-writeChannel:
			Send(conn, msg)
		}
	}
}
