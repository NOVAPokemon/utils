package api

import (
	"fmt"
	"github.com/NOVAPokemon/utils"
	"net/http"
)

const StatusPath = "/"
const StatusName = "STATUS"

const GET = "GET"

const StatusOnline = "online"

// Indicates if the server is online.
func Status(w http.ResponseWriter, _ *http.Request) {
	_, _ = fmt.Fprintln(w, StatusOnline)
}

var DefaultRoute = utils.Route{
	Name:        StatusName,
	Method:      GET,
	Pattern:     StatusPath,
	HandlerFunc: Status,
}