module github.com/NOVAPokemon/utils

go 1.13

require (
	github.com/bruno-anjos/archimedesHTTPClient v0.0.2
	github.com/dgrijalva/jwt-go v3.2.0+incompatible
	github.com/golang/geo v0.0.0-20200319012246-673a6f80352d
	github.com/gorilla/mux v1.7.4
	github.com/gorilla/websocket v1.4.2
	github.com/mitchellh/mapstructure v1.3.3
	github.com/pkg/errors v0.9.1
	github.com/prometheus/client_golang v1.6.0
	github.com/sirupsen/logrus v1.6.0
	github.com/stretchr/testify v1.4.0
	go.mongodb.org/mongo-driver v1.3.1
)

replace (
	github.com/bruno-anjos/archimedesHTTPClient v0.0.2 => ./../../bruno-anjos/archimedesHTTPClient
	github.com/bruno-anjos/cloud-edge-deployment v0.0.1 => ../../bruno-anjos/cloud-edge-deployment
)
