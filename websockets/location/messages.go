package location

import (
	"github.com/golang/geo/s2"

	"github.com/NOVAPokemon/utils"
	"github.com/NOVAPokemon/utils/items"
	ws "github.com/NOVAPokemon/utils/websockets"
)

const (
	UpdateLocation          = "UPDATE_LOCATION"
	UpdateLocationWithTiles = "UPDATE_LOCATION_WITH_TILES"
	Gyms                    = "GYMS"
	Pokemon                 = "POKEMON"
	CatchPokemon            = "CATCH_POKEMON"
	CatchPokemonResponse    = "CATCH_POKEMON_RESPONSE"
	ServersResponse         = "SERVERS_RESPONSE"
	CellsResponse           = "TILES_RESPONSE"
	Disconnect              = "DISCONNECT"
)

type UpdateLocationMessage struct {
	Location s2.LatLng
}

func (ulMsg UpdateLocationMessage) ConvertToWSMessage() *ws.WebsocketMsg {
	return ws.NewRequestMsg(UpdateLocation, ulMsg)
}

type UpdateLocationWithTilesMessage struct {
	CellsPerServer map[string]s2.CellUnion
}

func (ulwtMsg UpdateLocationWithTilesMessage) ConvertToWSMessage() *ws.WebsocketMsg {
	return ws.NewRequestMsg(UpdateLocationWithTiles, ulwtMsg)
}

type ServersMessage struct {
	Servers []string
}

func (serverMsg ServersMessage) ConvertToWSMessage(info ws.TrackedInfo) *ws.WebsocketMsg {
	return ws.NewReplyMsg(ServersResponse, serverMsg, info)
}

type CellsPerServerMessage struct {
	CellsPerServer map[string]s2.CellUnion
	OriginServer   string
}

func (tilesMsg CellsPerServerMessage) ConvertToWSMessage(info ws.TrackedInfo) *ws.WebsocketMsg {
	return ws.NewReplyMsg(CellsResponse, tilesMsg, info)
}

type GymsMessage struct {
	Gyms []utils.GymWithServer
}

func (gMsg GymsMessage) ConvertToWSMessage(info ws.TrackedInfo) *ws.WebsocketMsg {
	return ws.NewReplyMsg(Gyms, gMsg, info)
}

type PokemonMessage struct {
	Pokemon []utils.WildPokemonWithServer
}

func (pMsg PokemonMessage) ConvertToWSMessage(info ws.TrackedInfo) *ws.WebsocketMsg {
	return ws.NewReplyMsg(Pokemon, pMsg, info)
}

type CatchWildPokemonMessage struct {
	Pokeball    items.Item
	WildPokemon utils.WildPokemonWithServer
}

func (cMsg CatchWildPokemonMessage) ConvertToWSMessage() *ws.WebsocketMsg {
	return ws.NewRequestMsg(CatchPokemon, cMsg)
}

type CatchWildPokemonMessageResponse struct {
	Caught        bool
	PokemonTokens []string
	Error         string
}

func (cMsgResp CatchWildPokemonMessageResponse) ConvertToWSMessage(info ws.TrackedInfo) *ws.WebsocketMsg {
	return ws.NewReplyMsg(CatchPokemonResponse, cMsgResp, info)
}

type DisconnectMessage struct {
	Addr string
}

func (m DisconnectMessage) ConvertToWSMessage() *ws.WebsocketMsg {
	return ws.NewStandardMsg(Disconnect, m)
}
