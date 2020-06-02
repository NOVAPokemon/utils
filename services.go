package utils

import (
	"flag"
	"fmt"
	"math/rand"
	"net/http"
	"os"
	"time"

	"github.com/NOVAPokemon/utils/websockets"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	log "github.com/sirupsen/logrus"
)

const (
	AuthenticationPort = 8001 + iota
	BattlesPort
	GymPort
	LocationPort
	MicrotransactionsPort
	NotificationsPort
	StorePort
	TradesPort
	TrainersPort
)

const (
	AuthenticationEnvVar    = "AUTHENTICATION_URL"
	BattlesEnvVar           = "BATTLES_URL"
	GymEnvVar               = "GYM_URL"
	LocationEnvVar          = "LOCATION_URL"
	MicrotransactionsEnvVar = "MICROTRANSACTIONS_URL"
	NotificationsEnvVar     = "NOTIFICATIONS_URL"
	StoreEnvVar             = "STORE_URL"
	TradesEnvVar            = "TRADES_URL"
	TrainersEnvVar          = "TRAINERS_URL"

	HeadlessServiceNameEnvVar = "HEADLESS_SERVICE_NAME"
	HostnameEnvVar            = "HOSTNAME"
	MongoEnvVar               = "MONGODB_URL"

	KafkaEnvVar = "KAFKA_URL"

	Host      = "localhost"
	ServeHost = "0.0.0.0"
)

const (
	logDir = "/logs"
)

func StartServer(serviceName, host string, port int, routes Routes) {
	rand.Seed(time.Now().UnixNano())
	addr := fmt.Sprintf("%s:%d", host, port)
	r := NewRouter(routes)
	r.Handle("/metrics", promhttp.Handler())
	log.Infof("Starting %s server in port %d...\n", serviceName, port)
	log.Fatal(http.ListenAndServe(addr, r))
}

func CheckLogFlag(serviceName string) {
	logToStdout := GetLogFlag(serviceName)

	if !logToStdout {
		SetLogFile(serviceName)
	}
}

func GetLogFlag(serviceName string) bool {
	flag.Usage = func() {
		fmt.Println("Usage:")
		fmt.Printf("%s -l \n", serviceName)
	}
	var logToStdout bool
	flag.BoolVar(&logToStdout, "l", false, "log to stdout")
	flag.Parse()

	return logToStdout
}

func SetLogFile(serviceName string) {
	timestamp := websockets.MakeTimestamp()
	logFile, err := os.Create(fmt.Sprintf("%s/%s-%d", logDir, serviceName, timestamp))
	if err != nil {
		log.Fatal(err)
	}

	log.SetOutput(logFile)
}
