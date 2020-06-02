package battles

import (
	"fmt"
	"time"

	"github.com/NOVAPokemon/utils/pokemons"
	ws "github.com/NOVAPokemon/utils/websockets"
	"github.com/gorilla/websocket"
	log "github.com/sirupsen/logrus"
)

func HandleUseItem(useItemMessage *UseItemMessage, issuer *TrainerBattleStatus, issuerChan chan ws.GenericMsg,
	cooldownDuration time.Duration) bool {
	if issuer.Cooldown {
		issuerChan <- ws.GenericMsg{
			MsgType: websocket.TextMessage,
			Data: []byte(ws.ErrorMessage{
				Info:  fmt.Sprintf(ErrorCooldown.Error()),
				Fatal: false,
			}.SerializeToWSMessage().Serialize()),
		}

		return false
	}

	itemId := useItemMessage.ItemId
	item, ok := issuer.TrainerItems[itemId]
	if !ok {
		issuerChan <- ws.GenericMsg{
			MsgType: websocket.TextMessage,
			Data: []byte(ws.ErrorMessage{
				Info:  fmt.Sprintf(ErrorInvalidItemSelected.Error()),
				Fatal: false,
			}.SerializeToWSMessage().Serialize()),
		}

		return false
	}

	if !item.Effect.Appliable {
		issuerChan <- ws.GenericMsg{
			MsgType: websocket.TextMessage,
			Data: []byte(ws.ErrorMessage{
				Info:  fmt.Sprintf(ErrorItemNotAppliable.Error()),
				Fatal: false,
			}.SerializeToWSMessage().Serialize()),
		}
	}

	err := item.Apply(issuer.SelectedPokemon)
	if err != nil {
		issuerChan <- ws.GenericMsg{
			MsgType: websocket.TextMessage,
			Data: []byte(ws.ErrorMessage{
				Info:  fmt.Sprintf(err.Error()),
				Fatal: false,
			}.SerializeToWSMessage().Serialize()),
		}
	}

	issuer.CdTimer.Reset(cooldownDuration)
	issuer.Cooldown = true

	issuer.UsedItems[item.Id.Hex()] = item
	delete(issuer.TrainerItems, item.Id.Hex())
	UpdateTrainerPokemon(useItemMessage.TrackedMessage, *issuer.SelectedPokemon, issuerChan, true)
	issuerChan <- ws.GenericMsg{
		MsgType: websocket.TextMessage,
		Data: []byte(RemoveItemMessage{
			ItemId: itemId,
		}.SerializeToWSMessage().Serialize()),
	}
	return true
}

func HandleSelectPokemon(selectedPokemonMsg *SelectPokemonMessage, issuer *TrainerBattleStatus,
	issuerChan chan ws.GenericMsg) bool {
	selectedPokemonId := selectedPokemonMsg.PokemonId
	pokemon, ok := issuer.TrainerPokemons[selectedPokemonId]
	if !ok {
		issuerChan <- ws.GenericMsg{
			MsgType: websocket.TextMessage,
			Data: []byte(ws.ErrorMessage{
				Info:  fmt.Sprintf(ErrorInvalidPokemonSelected.Error()),
				Fatal: false,
			}.SerializeToWSMessage().Serialize()),
		}
		return false
	}

	if pokemon.HP <= 0 {
		issuerChan <- ws.GenericMsg{
			MsgType: websocket.TextMessage,
			Data: []byte(ws.ErrorMessage{
				Info:  fmt.Sprintf(ErrorPokemonNoHP.Error()),
				Fatal: false,
			}.SerializeToWSMessage().Serialize()),
		}
		return false
	}
	issuer.SelectedPokemon = pokemon
	log.Info("Changed selected pokemon")
	UpdateTrainerPokemon(selectedPokemonMsg.TrackedMessage, *issuer.SelectedPokemon, issuerChan, true)
	return true
}

func HandleDefendMove(issuer *TrainerBattleStatus, issuerChan chan ws.GenericMsg, cooldownDuration time.Duration) {
	// if the pokemon is dead, player must select a new pokemon
	if issuer.SelectedPokemon.HP == 0 {
		issuerChan <- ws.GenericMsg{
			MsgType: websocket.TextMessage,
			Data: []byte(ws.ErrorMessage{
				Info:  fmt.Sprintf(ErrorPokemonNoHP.Error()),
				Fatal: false,
			}.SerializeToWSMessage().Serialize()),
		}
		return
	}

	// if player has moved recently and is in Cooldown, discard move
	if issuer.Cooldown {
		issuerChan <- ws.GenericMsg{
			MsgType: websocket.TextMessage,
			Data: []byte(ws.ErrorMessage{
				Info:  fmt.Sprintf(ErrorCooldown.Error()),
				Fatal: false,
			}.SerializeToWSMessage().Serialize()),
		}
		return
	}
	issuer.CdTimer.Reset(cooldownDuration)
	issuer.Cooldown = true

	// process Defending move: update both players and setup a Cooldown
	issuer.Defending = true
	issuerChan <- ws.GenericMsg{
		MsgType: websocket.TextMessage,
		Data: []byte(StatusMessage{
			Message: StatusDefended,
		}.SerializeToWSMessage().Serialize()),
	}
	return
}

func HandleAttackMove(issuer *TrainerBattleStatus, issuerChan chan ws.GenericMsg, defending bool,
	otherPokemon *pokemons.Pokemon, cooldownDuration time.Duration) bool {
	if issuer.SelectedPokemon.HP == 0 {
		issuerChan <- ws.GenericMsg{
			MsgType: websocket.TextMessage,
			Data: []byte(ws.ErrorMessage{
				Info:  fmt.Sprintf(ErrorPokemonNoHP.Error()),
				Fatal: false,
			}.SerializeToWSMessage().Serialize()),
		}
		return false
	}

	// if player has moved recently and is in Cooldown, discard move
	if issuer.Cooldown {
		issuerChan <- ws.GenericMsg{
			MsgType: websocket.TextMessage,
			Data: []byte(ws.ErrorMessage{
				Info:  fmt.Sprintf(ErrorCooldown.Error()),
				Fatal: false,
			}.SerializeToWSMessage().Serialize()),
		}
		return false
	}

	issuer.CdTimer.Reset(cooldownDuration)
	issuer.Cooldown = true
	hpChanged := ApplyAttackMove(issuer.SelectedPokemon, otherPokemon, defending)

	return hpChanged
}

func ApplyAttackMove(issuerPokemon *pokemons.Pokemon, otherPokemon *pokemons.Pokemon, defending bool) bool {
	if defending {
		return false
	} else {
		otherPokemon.HP -= issuerPokemon.Damage
		if otherPokemon.HP < 0 {
			otherPokemon.HP = 0
		}
		return true
	}
}

func UpdateTrainerPokemon(trackedMsg ws.TrackedMessage, pokemon pokemons.Pokemon, channel chan ws.GenericMsg, owner bool) {
	channel <- ws.GenericMsg{
		MsgType: websocket.TextMessage,
		Data: []byte(UpdatePokemonMessage{
			Owner:          owner,
			Pokemon:        pokemon,
			TrackedMessage: trackedMsg,
		}.SerializeToWSMessage().Serialize()),
	}
}
