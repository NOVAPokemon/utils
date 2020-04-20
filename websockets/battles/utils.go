package battles

import (
	"fmt"
	"github.com/NOVAPokemon/utils/pokemons"
	"github.com/NOVAPokemon/utils/websockets"
	log "github.com/sirupsen/logrus"
)

func HandleUseItem(useItemMessage *UseItemMessage, issuer *TrainerBattleStatus, issuerChan chan *string) bool {
	if issuer.Cooldown {
		websockets.SendMessage(
			*ErrorMessage{
				Info:  fmt.Sprintf(ErrCooldown.Error()),
				Fatal: false,
			}.SerializeToWSMessage(), issuerChan)
		return false
	}

	itemId := useItemMessage.ItemId
	item, ok := issuer.TrainerItems[itemId]
	if !ok {
		websockets.SendMessage(
			*ErrorMessage{
				Info:  fmt.Sprintf(ErrInvalidItemSelected.Error()),
				Fatal: false,
			}.SerializeToWSMessage(), issuerChan)
		return false
	}

	if !item.Effect.Appliable {
		websockets.SendMessage(
			*ErrorMessage{
				Info:  fmt.Sprintf(ErrItemNotAppliable.Error()),
				Fatal: false,
			}.SerializeToWSMessage(), issuerChan)
		return false
	}

	err := item.Apply(issuer.SelectedPokemon)
	if err != nil {
		websockets.SendMessage(
			*ErrorMessage{
				Info:  fmt.Sprintf(err.Error()),
				Fatal: false,
			}.SerializeToWSMessage(), issuerChan)
	}

	issuer.CdTimer.Reset(DefaultCooldown)
	issuer.Cooldown = true

	issuer.UsedItems[item.Id.Hex()] = item
	delete(issuer.TrainerItems, item.Id.Hex())
	UpdateTrainerPokemon(useItemMessage.TrackedMessage, *issuer.SelectedPokemon, issuerChan, true)
	websockets.SendMessage(
		*RemoveItemMessage{
			ItemId: itemId,
		}.SerializeToWSMessage(), issuerChan)

	return true
}

func HandleSelectPokemon(selectedPokemonMsg *SelectPokemonMessage, issuer *TrainerBattleStatus, issuerChan chan *string) bool {

	selectedPokemonId := selectedPokemonMsg.PokemonId
	pokemon, ok := issuer.TrainerPokemons[selectedPokemonId]

	if !ok {
		websockets.SendMessage(
			*ErrorMessage{
				Info:  fmt.Sprintf(ErrInvalidPokemonSelected.Error()),
				Fatal: false,
			}.SerializeToWSMessage(), issuerChan)
		return false
	}
	if pokemon.HP <= 0 {
		websockets.SendMessage(
			*ErrorMessage{
				Info:  fmt.Sprintf(ErrPokemonNoHP.Error()),
				Fatal: false,
			}.SerializeToWSMessage(), issuerChan)
		return false
	}
	issuer.SelectedPokemon = pokemon
	log.Info("Changed selected pokemon")
	UpdateTrainerPokemon(selectedPokemonMsg.TrackedMessage, *issuer.SelectedPokemon, issuerChan, true)
	return true
}

func HandleDefendMove(issuer *TrainerBattleStatus, issuerChan chan *string) {

	// if the pokemon is dead, player must select a new pokemon
	if issuer.SelectedPokemon.HP == 0 {
		websockets.SendMessage(
			*ErrorMessage{
				Info:  fmt.Sprintf(ErrPokemonNoHP.Error()),
				Fatal: false,
			}.SerializeToWSMessage(), issuerChan)
		return
	}

	// if player has moved recently and is in Cooldown, discard move
	if issuer.Cooldown {
		websockets.SendMessage(
			*ErrorMessage{
				Info:  fmt.Sprintf(ErrCooldown.Error()),
				Fatal: false,
			}.SerializeToWSMessage(), issuerChan)
		return
	}
	issuer.CdTimer.Reset(DefaultCooldown)
	issuer.Cooldown = true

	// process Defending move: update both players and setup a Cooldown
	issuer.Defending = true
	websockets.SendMessage(
		*StatusMessage{
			Message: StatusDefended,
		}.SerializeToWSMessage(), issuerChan)
	return
}

func HandleAttackMove(issuer *TrainerBattleStatus, issuerChan chan *string, defending bool, otherPokemon *pokemons.Pokemon) bool {
	if issuer.SelectedPokemon.HP == 0 {
		websockets.SendMessage(
			*ErrorMessage{
				Info:  fmt.Sprintf(ErrPokemonNoHP.Error()),
				Fatal: false,
			}.SerializeToWSMessage(), issuerChan)
		return false
	}

	// if player has moved recently and is in Cooldown, discard move
	if issuer.Cooldown {
		websockets.SendMessage(
			*ErrorMessage{
				Info:  fmt.Sprintf(ErrCooldown.Error()),
				Fatal: false,
			}.SerializeToWSMessage(), issuerChan)
		return false
	}

	issuer.CdTimer.Reset(DefaultCooldown)
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

func UpdateTrainerPokemon(trackedMsg websockets.TrackedMessage, pokemon pokemons.Pokemon, channel chan *string, owner bool) {
	websockets.SendMessage(
		*UpdatePokemonMessage{
			Owner:          owner,
			Pokemon:        pokemon,
			TrackedMessage: trackedMsg,
		}.SerializeToWSMessage(), channel)
}
