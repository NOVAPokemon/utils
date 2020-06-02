package api

import (
	"fmt"
	"github.com/NOVAPokemon/utils"
	"net/http"
)

const (
	StatusName = "STATUS"

	GET = "GET"

	StatusOnline = "online"
)

// Indicates if the server is online.
func Status(w http.ResponseWriter, _ *http.Request) {
	_, _ = fmt.Fprintln(w, StatusOnline)
}

func GenStatusRoute(serverName string) utils.Route {
	return utils.Route{
		Name:        StatusName,
		Method:      GET,
		Pattern:     fmt.Sprintf("/%s/status", serverName),
		HandlerFunc: Status,
	}
}
