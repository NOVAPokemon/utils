package utils

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"
	"os"
	"time"

	"github.com/NOVAPokemon/utils/websockets"
	"github.com/NOVAPokemon/utils/websockets/comms_manager"
	"github.com/golang/geo/s2"
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
	IngressEnvVar           = "INGRESS_URL"
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
	logDir                      = "/logs"
	DefaultDelayConfigFilename  = "delays_config.json"
	DefaultClientDelaysFilename = "client_delays.json"
)

type (
	OptionalConfigs struct {
		CellID s2.CellID
	}
)

func StartServer(serviceName, host string, port int, routes Routes, manager websockets.CommunicationManager) {
	rand.Seed(time.Now().UnixNano())
	addr := fmt.Sprintf("%s:%d", host, port)

	r := NewRouter(routes)
	r.Handle("/metrics", promhttp.Handler())

	r.Use(manager.HTTPRequestInterceptor)

	log.Infof("Starting %s server in port %d...\n", serviceName, port)
	server := http.Server{
		Addr:              addr,
		Handler:           r,
		TLSConfig:         nil,
		ReadTimeout:       websockets.Timeout,
		ReadHeaderTimeout: websockets.Timeout,
		WriteTimeout:      websockets.Timeout,
		IdleTimeout:       websockets.Timeout,
	}
	log.Fatal(server.ListenAndServe())
}

func SetLogFile(serviceName string) {
	timestamp := websockets.MakeTimestamp()
	logFile, err := os.Create(fmt.Sprintf("%s/%s-%d", logDir, serviceName, timestamp))
	if err != nil {
		log.Fatal(err)
	}

	log.SetOutput(logFile)
}

func CreateDefaultDelayedManager(isClient bool, optConfigs *OptionalConfigs) websockets.CommunicationManager {
	var (
		delaysConfig string
		clientDelays string
		ok           bool
	)

	if delaysConfig, ok = os.LookupEnv("DELAYS_CONFIG"); !ok {
		delaysConfig = DefaultDelayConfigFilename
	}

	if clientDelays, ok = os.LookupEnv("CLIENT_DELAYS"); !ok {
		clientDelays = DefaultClientDelaysFilename
	}

	return createDelayedCommunicationManager(delaysConfig, clientDelays, isClient,
		optConfigs)
}

func createDelayedCommunicationManager(delayedCommsFilename, clientDelaysFilename string,
	isClient bool, optConfigs *OptionalConfigs) websockets.CommunicationManager {
	log.Info("using DELAYED communication manager")

	if optConfigs == nil {
		panic("optConfigs is nil and s2delayed needs cellID")
	}

	if optConfigs.CellID.ToToken() == "X" {
		log.Panicf("invalid cellID %s", optConfigs.CellID.ToToken())
	}

	delaysConfig := getDelayedConfig(delayedCommsFilename)

	return comms_manager.NewS2DelayedCommsManager(optConfigs.CellID, delaysConfig, isClient)
}

func CreateDefaultCommunicationManager() websockets.CommunicationManager {
	log.Info("using DEFAULT communication manager")
	return &comms_manager.DefaultCommsManager{}
}

func getDelayedConfig(delayedCommsFilename string) *comms_manager.DelaysMatrixType {
	file, err := ioutil.ReadFile(delayedCommsFilename)
	if err != nil {
		panic(fmt.Sprintf("could not read %s: %s", delayedCommsFilename, err))
	}

	var delaysMatrix comms_manager.DelaysMatrixType
	err = json.Unmarshal(file, &delaysMatrix)
	if err != nil {
		panic(err)
	}

	return &delaysMatrix
}

func getClientDelays(clientDelaysFilename string) *comms_manager.ClientDelays {
	file, err := ioutil.ReadFile(clientDelaysFilename)
	if err != nil {
		panic(fmt.Sprintf("could not read %s: %s", clientDelaysFilename, err))
	}

	var clientDelays comms_manager.ClientDelays
	err = json.Unmarshal(file, &clientDelays)
	if err != nil {
		panic(err)
	}

	return &clientDelays
}

type Flags struct {
	LogToStdout  *bool
	DelayedComms *bool
}

func setLogFlag() *bool {
	var logToStdout bool
	flag.BoolVar(&logToStdout, "l", false, "log to stdout")
	return &logToStdout
}

func SetDelayedFlag() *bool {
	var delayed bool
	flag.BoolVar(&delayed, "d", false, "add delays to communication")
	return &delayed
}

func ParseFlags(serviceName string) Flags {
	flag.Usage = func() {
		fmt.Println("Usage:")
		fmt.Printf("%s -l -d\n", serviceName)
	}

	logToStdOut := setLogFlag()
	delayedComms := SetDelayedFlag()

	flag.Parse()

	return Flags{
		LogToStdout:  logToStdOut,
		DelayedComms: delayedComms,
	}
}

func GetLocationTag(filename, serverName string) string {
	file, err := ioutil.ReadFile(filename)
	if err != nil {
		panic(fmt.Sprintf("error getting location tags while loading file %s", filename))
	}

	var locationTags map[string]string
	err = json.Unmarshal(file, &locationTags)
	if err != nil {
		panic(err)
	}

	if locationTag, ok := locationTags[serverName]; !ok {
		panic(fmt.Sprintf("could not find location tag for servername %s", serverName))
	} else {
		return locationTag
	}
}
