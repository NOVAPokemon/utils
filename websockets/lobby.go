package websockets

import (
	"encoding/json"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	log "github.com/sirupsen/logrus"
	dry "github.com/ungerik/go-dry"
)

const (
	PongWait   = 2 * time.Second
	PingPeriod = (PongWait * 9) / 10
)

type DebugMutex struct {
	id string
	m  sync.Mutex
}

func (d *DebugMutex) Lock() {
	log.Infof("%s.Lock() %s\n", d.id, dry.StackTraceLine(3))
	d.m.Lock()
}

func (d *DebugMutex) Unlock() {
	log.Infof("%s.Unlock() %s\n", d.id, dry.StackTraceLine(3))
	d.m.Unlock()
}

// Lobby maintains the connections from both trainers and the status of the battle
type Lobby struct {
	Id              string
	changeLobbyLock *DebugMutex
	TrainersJoined  int
	Started         chan struct{}
	Finished        chan struct{}
	Capacity        int

	TrainerUsernames      []string
	TrainerInChannels     []chan *WebsocketMsg
	DoneListeningFromConn []chan interface{}
	DoneWritingToConn     []chan interface{}
	TrainerOutChannels    []chan *WebsocketMsg
	trainerConnections    []*websocket.Conn
	finishOnce            sync.Once

	StartTrackInfo *TrackedInfo
}

func NewLobby(id string, capacity int, startTrackInfo *TrackedInfo) *Lobby {
	return &Lobby{
		Capacity:              capacity,
		Id:                    id,
		TrainersJoined:        0,
		TrainerUsernames:      make([]string, capacity),
		trainerConnections:    make([]*websocket.Conn, capacity),
		TrainerInChannels:     make([]chan *WebsocketMsg, capacity),
		DoneListeningFromConn: make([]chan interface{}, capacity),
		TrainerOutChannels:    make([]chan *WebsocketMsg, capacity),
		DoneWritingToConn:     make([]chan interface{}, capacity),
		Started:               make(chan struct{}),
		Finished:              make(chan struct{}),
		changeLobbyLock:       &DebugMutex{id: id},
		finishOnce:            sync.Once{},
		StartTrackInfo:        startTrackInfo,
	}
}

func AddTrainer(lobby *Lobby, username string, trainerConn *websocket.Conn,
	commsManager CommunicationManager) (int, error) {
	lobby.changeLobbyLock.Lock()
	defer lobby.changeLobbyLock.Unlock()

	if lobby.TrainersJoined >= lobby.Capacity {
		return -1, NewLobbyIsFullError(lobby.Id)
	}

	select {
	case <-lobby.Started:
		return -1, NewLobbyStartedError(lobby.Id)
	case <-lobby.Finished:
		return -1, NewLobbyFinishedError(lobby.Id)
	default:
		trainerNum := lobby.TrainersJoined
		trainerChanIn := make(chan *WebsocketMsg)
		trainerChanOut := make(chan *WebsocketMsg)

		lobby.TrainerUsernames[trainerNum] = username
		lobby.TrainerInChannels[trainerNum] = trainerChanIn
		lobby.TrainerOutChannels[trainerNum] = trainerChanOut
		lobby.trainerConnections[trainerNum] = trainerConn
		lobby.DoneListeningFromConn[trainerNum] = RecvFromConnToChann(lobby, trainerNum, commsManager)
		lobby.DoneWritingToConn[trainerNum] = sendFromChanToConn(lobby, trainerNum, commsManager)
		lobby.TrainersJoined++
		return lobby.TrainersJoined, nil
	}
}

func sendFromChanToConn(lobby *Lobby, trainerNum int, writer CommunicationManager) (done chan interface{}) {
	done = make(chan interface{})
	go func() {
		pingTicker := time.NewTicker(PingPeriod)
		conn := lobby.trainerConnections[trainerNum]
		outChannel := lobby.TrainerOutChannels[trainerNum]
		conn.SetPongHandler(func(_ string) error {
			return conn.SetReadDeadline(time.Now().Add(PongWait))
		})

		defer close(done)

		for {
			select {
			case <-pingTicker.C:
				err := writer.WriteGenericMessageToConn(conn, NewControlMsg(websocket.PingMessage))
				if err != nil {
					log.Warn(err)
					return
				}
			case msg, ok := <-outChannel:
				if !ok {
					continue
				}
				err := writer.WriteGenericMessageToConn(conn, msg)
				if err != nil {
					log.Warn(err)
					return
				}
			case <-lobby.Finished:
				log.Info("Send routine finishing")
				return
			}
		}
	}()
	return done
}

func RecvFromConnToChann(lobby *Lobby, trainerNum int, manager CommunicationManager) (done chan interface{}) {
	done = make(chan interface{})

	messagesQueue := make(chan (<-chan *WebsocketMsg), 10)

	inChannel := lobby.TrainerInChannels[trainerNum]

	cancel := make(chan struct{})

	go func() {
		log.Infof("(%s, %s) started message queue routine",
			lobby.Id, lobby.TrainerUsernames[trainerNum])
		defer func() {
			log.Infof("(%s, %s) finished message queue routine",
				lobby.Id, lobby.TrainerUsernames[trainerNum])
			close(done)
		}()
		for {
			select {
			case <-lobby.Finished:
				log.Infof("(%s, %s) could not send message because finish channel ended meanwhile in message queue",
					lobby.Id, lobby.TrainerUsernames[trainerNum])
				return
			case <-cancel:
				log.Infof("(%s, %s) triggered cancel",
					lobby.Id, lobby.TrainerUsernames[trainerNum])
				return
			case chanToWait := <-messagesQueue:
				log.Infof("(%s, %s) waiting for message",
					lobby.Id, lobby.TrainerUsernames[trainerNum])
				msg := <-chanToWait
				log.Infof("(%s, %s) got message %+v",
					lobby.Id, lobby.TrainerUsernames[trainerNum], msg)

				select {
				case inChannel <- msg:
					log.Infof("(%s, %s) wrote message %+v",
						lobby.Id, lobby.TrainerUsernames[trainerNum], msg)
				case <-lobby.Finished:
					log.Infof("(%s, %s) lobby finished in the meanwhile",
						lobby.Id, lobby.TrainerUsernames[trainerNum])
				}
			}
		}
	}()

	go func() {
		conn := lobby.trainerConnections[trainerNum]

		defer close(cancel)
		for {
			msgChan, err := manager.ReadMessageFromConn(conn)
			if err != nil {
				log.Infof("(%s, %s) Receive routine finishing because connection was closed",
					lobby.Id, lobby.TrainerUsernames[trainerNum])
				return
			}

			select {
			case <-lobby.Finished:
				log.Infof("(%s, %s) could not send message because finish channel ended meanwhile",
					lobby.Id, lobby.TrainerUsernames[trainerNum])
				return
			default:
				if msgChan != nil {
					messagesQueue <- msgChan
				}
			}
		}
	}()
	return done
}

func StartLobby(lobby *Lobby) {
	close(lobby.Started)
}

func FinishLobby(lobby *Lobby) {
	lobby.finishOnce.Do(func() {
		log.Infof("finishing lobby %s", lobby.Id)
		lobby.changeLobbyLock.Lock()
		defer lobby.changeLobbyLock.Unlock()
		close(lobby.Finished)
		for i := 0; i < lobby.TrainersJoined; i++ {
			if err := lobby.trainerConnections[i].Close(); err != nil {
				log.Error(err)
			}
			<-lobby.DoneWritingToConn[i]
			<-lobby.DoneListeningFromConn[i]
			close(lobby.TrainerOutChannels[i])
			close(lobby.TrainerInChannels[i])
		}
	})
}

func GetTrainersJoined(lobby *Lobby) int {
	lobby.changeLobbyLock.Lock()
	defer lobby.changeLobbyLock.Unlock()
	return lobby.TrainersJoined
}

func ParseContent(msgData []byte) *WebsocketMsgContent {
	toReturn := &WebsocketMsgContent{}

	if err := json.Unmarshal(msgData, toReturn); err != nil {
		log.Error(msgData)
		panic(wrapMsgParsingError(err))
	}

	return toReturn
}
