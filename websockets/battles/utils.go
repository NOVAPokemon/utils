package battles

import (
	"fmt"
	"time"

	"github.com/NOVAPokemon/utils/pokemons"
	ws "github.com/NOVAPokemon/utils/websockets"
	log "github.com/sirupsen/logrus"
)

func HandleUseItem(info *ws.TrackedInfo, useItemMsg *UseItemMessage, issuer *TrainerBattleStatus,
	issuerChan chan *ws.WebsocketMsg, cooldownDuration time.Duration) bool {

	if issuer.Cooldown {
		issuerChan <- ErrorBattleMessage{
			Info:  fmt.Sprintf(ErrorCooldown.Error()),
			Fatal: false,
		}.ConvertToWSMessage(*info)

		return false
	}

	itemId := useItemMsg.ItemId
	item, ok := issuer.TrainerItems[itemId]
	if !ok {
		issuerChan <- ErrorBattleMessage{
			Info:  fmt.Sprintf(ErrorInvalidItemSelected.Error()),
			Fatal: false,
		}.ConvertToWSMessage(*info)

		return false
	}

	if !item.Effect.Appliable {
		issuerChan <- ErrorBattleMessage{
			Info:  fmt.Sprintf(ErrorItemNotAppliable.Error()),
			Fatal: false,
		}.ConvertToWSMessage(*info)
	}

	err := item.Apply(issuer.SelectedPokemon)
	if err != nil {
		issuerChan <- ErrorBattleMessage{
			Info:  fmt.Sprintf(err.Error()),
			Fatal: false,
		}.ConvertToWSMessage(*info)
	}

	issuer.CdTimer.Reset(cooldownDuration)
	issuer.Cooldown = true

	issuer.UsedItems[item.Id] = item
	delete(issuer.TrainerItems, item.Id)
	UpdateTrainerPokemon(info, *issuer.SelectedPokemon, issuerChan, true)
	issuerChan <- RemoveItemMessage{
		ItemId: itemId,
	}.ConvertToWSMessage(*info)
	return true
}

func HandleSelectPokemon(info *ws.TrackedInfo, selectedPokemonMsg *SelectPokemonMessage, issuer *TrainerBattleStatus,
	issuerChan chan *ws.WebsocketMsg) bool {

	selectedPokemonId := selectedPokemonMsg.PokemonId
	pokemon, ok := issuer.TrainerPokemons[selectedPokemonId]
	if !ok {
		errMsg := ErrorBattleMessage{
			Info:  fmt.Sprintf(ErrorInvalidPokemonSelected.Error()),
			Fatal: false,
		}

		issuerChan <- errMsg.ConvertToWSMessage(*info)
		return false
	}

	if pokemon.HP <= 0 {
		errMsg := ErrorBattleMessage{
			Info:  fmt.Sprintf(ErrorPokemonNoHP.Error()),
			Fatal: false,
		}

		issuerChan <- errMsg.ConvertToWSMessage(*info)
		return false
	}

	issuer.SelectedPokemon = pokemon
	log.Info("Changed selected pokemon")
	UpdateTrainerPokemon(info, *issuer.SelectedPokemon, issuerChan, true)
	return true
}

func HandleDefendMove(info *ws.TrackedInfo, issuer *TrainerBattleStatus, issuerChan chan *ws.WebsocketMsg,
	cooldownDuration time.Duration) {
	// if the pokemon is dead, player must select a new pokemon
	if issuer.SelectedPokemon.HP == 0 {
		issuerChan <- ErrorBattleMessage{
			Info:  fmt.Sprintf(ErrorPokemonNoHP.Error()),
			Fatal: false,
		}.ConvertToWSMessage(*info)
		return
	}

	// if player has moved recently and is in Cooldown, discard move
	if issuer.Cooldown {
		issuerChan <- ErrorBattleMessage{
			Info:  fmt.Sprintf(ErrorCooldown.Error()),
			Fatal: false,
		}.ConvertToWSMessage(*info)
		return
	}
	issuer.CdTimer.Reset(cooldownDuration)
	issuer.Cooldown = true

	// process Defending move: update both players and setup a Cooldown
	issuer.Defending = true
	issuerChan <- StatusMessage{
		Message: StatusDefending,
	}.ConvertToWSMessage(*info)
}

func HandleAttackMove(info *ws.TrackedInfo, issuer *TrainerBattleStatus, issuerChan chan *ws.WebsocketMsg,
	defending bool, otherPokemon *pokemons.Pokemon, cooldownDuration time.Duration) bool {
	if issuer.SelectedPokemon.HP == 0 {
		issuerChan <- ErrorBattleMessage{
			Info:  fmt.Sprintf(ErrorPokemonNoHP.Error()),
			Fatal: false,
		}.ConvertToWSMessage(*info)
		return false
	}

	// if player has moved recently and is in Cooldown, discard move
	if issuer.Cooldown {
		issuerChan <- ErrorBattleMessage{
			Info:  fmt.Sprintf(ErrorCooldown.Error()),
			Fatal: false,
		}.ConvertToWSMessage(*info)
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

func UpdateTrainerPokemon(trackedInfo *ws.TrackedInfo, pokemon pokemons.Pokemon, channel chan *ws.WebsocketMsg,
	owner bool) {
	var wsMsg *ws.WebsocketMsg
	if trackedInfo != nil {
		wsMsg = UpdatePokemonMessage{
			Owner:   owner,
			Pokemon: pokemon,
		}.ConvertToWSMessageWithInfo(*trackedInfo)
	} else {
		wsMsg = UpdatePokemonMessage{
			Owner:   owner,
			Pokemon: pokemon,
		}.ConvertToWSMessage()
	}

	channel <- wsMsg
}
