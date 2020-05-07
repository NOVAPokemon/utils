package battles

import (
	"fmt"
	"github.com/NOVAPokemon/utils/pokemons"
	ws "github.com/NOVAPokemon/utils/websockets"
	log "github.com/sirupsen/logrus"
	"time"
)

func HandleUseItem(useItemMessage *UseItemMessage, issuer *TrainerBattleStatus, issuerChan chan *string,
	cooldownDuration time.Duration) bool {
	if issuer.Cooldown {
		ws.SendMessage(
			*ws.ErrorMessage{
				Info:  fmt.Sprintf(ErrorCooldown.Error()),
				Fatal: false,
			}.SerializeToWSMessage(), issuerChan)

		return false
	}

	itemId := useItemMessage.ItemId
	item, ok := issuer.TrainerItems[itemId]
	if !ok {
		ws.SendMessage(
			*ws.ErrorMessage{
				Info:  fmt.Sprintf(ErrorInvalidItemSelected.Error()),
				Fatal: false,
			}.SerializeToWSMessage(), issuerChan)

		return false
	}

	if !item.Effect.Appliable {
		ws.SendMessage(
			*ws.ErrorMessage{
				Info:  fmt.Sprintf(ErrorItemNotAppliable.Error()),
				Fatal: false,
			}.SerializeToWSMessage(), issuerChan)
		return false
	}

	err := item.Apply(issuer.SelectedPokemon)
	if err != nil {
		ws.SendMessage(
			*ws.ErrorMessage{
				Info:  fmt.Sprintf(err.Error()),
				Fatal: false,
			}.SerializeToWSMessage(), issuerChan)
	}

	issuer.CdTimer.Reset(cooldownDuration)
	issuer.Cooldown = true

	issuer.UsedItems[item.Id.Hex()] = item
	delete(issuer.TrainerItems, item.Id.Hex())
	UpdateTrainerPokemon(useItemMessage.TrackedMessage, *issuer.SelectedPokemon, issuerChan, true)
	ws.SendMessage(
		*RemoveItemMessage{
			ItemId: itemId,
		}.SerializeToWSMessage(), issuerChan)

	return true
}

func HandleSelectPokemon(selectedPokemonMsg *SelectPokemonMessage, issuer *TrainerBattleStatus,
	issuerChan chan *string) bool {
	selectedPokemonId := selectedPokemonMsg.PokemonId
	pokemon, ok := issuer.TrainerPokemons[selectedPokemonId]
	if !ok {
		ws.SendMessage(
			*ws.ErrorMessage{
				Info:  fmt.Sprintf(ErrorInvalidPokemonSelected.Error()),
				Fatal: false,
			}.SerializeToWSMessage(), issuerChan)

		return false
	}

	if pokemon.HP <= 0 {
		ws.SendMessage(
			*ws.ErrorMessage{
				Info:  fmt.Sprintf(ErrorPokemonNoHP.Error()),
				Fatal: false,
			}.SerializeToWSMessage(), issuerChan)
		return false
	}
	issuer.SelectedPokemon = pokemon
	log.Info("Changed selected pokemon")
	UpdateTrainerPokemon(selectedPokemonMsg.TrackedMessage, *issuer.SelectedPokemon, issuerChan, true)
	return true
}

func HandleDefendMove(issuer *TrainerBattleStatus, issuerChan chan *string, cooldownDuration time.Duration) {
	// if the pokemon is dead, player must select a new pokemon
	if issuer.SelectedPokemon.HP == 0 {
		ws.SendMessage(
			*ws.ErrorMessage{
				Info:  fmt.Sprintf(ErrorPokemonNoHP.Error()),
				Fatal: false,
			}.SerializeToWSMessage(), issuerChan)
		return
	}

	// if player has moved recently and is in Cooldown, discard move
	if issuer.Cooldown {
		ws.SendMessage(
			*ws.ErrorMessage{
				Info:  fmt.Sprintf(ErrorCooldown.Error()),
				Fatal: false,
			}.SerializeToWSMessage(), issuerChan)
		return
	}
	issuer.CdTimer.Reset(cooldownDuration)
	issuer.Cooldown = true

	// process Defending move: update both players and setup a Cooldown
	issuer.Defending = true
	ws.SendMessage(
		*StatusMessage{
			Message: StatusDefended,
		}.SerializeToWSMessage(), issuerChan)
	return
}

func HandleAttackMove(issuer *TrainerBattleStatus, issuerChan chan *string, defending bool,
	otherPokemon *pokemons.Pokemon, cooldownDuration time.Duration) bool {
	if issuer.SelectedPokemon.HP == 0 {
		ws.SendMessage(
			*ws.ErrorMessage{
				Info:  fmt.Sprintf(ErrorPokemonNoHP.Error()),
				Fatal: false,
			}.SerializeToWSMessage(), issuerChan)
		return false
	}

	// if player has moved recently and is in Cooldown, discard move
	if issuer.Cooldown {
		ws.SendMessage(
			*ws.ErrorMessage{
				Info:  fmt.Sprintf(ErrorCooldown.Error()),
				Fatal: false,
			}.SerializeToWSMessage(), issuerChan)
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

func UpdateTrainerPokemon(trackedMsg ws.TrackedMessage, pokemon pokemons.Pokemon, channel chan *string, owner bool) {
	ws.SendMessage(
		*UpdatePokemonMessage{
			Owner:          owner,
			Pokemon:        pokemon,
			TrackedMessage: trackedMsg,
		}.SerializeToWSMessage(), channel)
}
