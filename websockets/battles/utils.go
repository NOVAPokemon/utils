package battles

import (
	"fmt"
	"time"

	"github.com/NOVAPokemon/utils/pokemons"
	ws "github.com/NOVAPokemon/utils/websockets"
	log "github.com/sirupsen/logrus"
)

func HandleUseItem(useItemMessage *UseItemMessage, issuer *TrainerBattleStatus, issuerChan chan ws.Serializable,
	cooldownDuration time.Duration) bool {
	if issuer.Cooldown {
		issuerChan <- ws.ErrorMessage{
			Info:  fmt.Sprintf(ErrorCooldown.Error()),
			Fatal: false,
		}

		return false
	}

	itemId := useItemMessage.ItemId
	item, ok := issuer.TrainerItems[itemId]
	if !ok {
		issuerChan <- ws.ErrorMessage{
			Info:  fmt.Sprintf(ErrorInvalidItemSelected.Error()),
			Fatal: false,
		}

		return false
	}

	if !item.Effect.Appliable {
		issuerChan <- ws.ErrorMessage{
			Info:  fmt.Sprintf(ErrorItemNotAppliable.Error()),
			Fatal: false,
		}
	}

	err := item.Apply(issuer.SelectedPokemon)
	if err != nil {
		issuerChan <- ws.ErrorMessage{
			Info:  fmt.Sprintf(err.Error()),
			Fatal: false,
		}
	}

	issuer.CdTimer.Reset(cooldownDuration)
	issuer.Cooldown = true

	issuer.UsedItems[item.Id.Hex()] = item
	delete(issuer.TrainerItems, item.Id.Hex())
	UpdateTrainerPokemon(useItemMessage.TrackedMessage, *issuer.SelectedPokemon, issuerChan, true)
	issuerChan <- RemoveItemMessage{
		ItemId: itemId,
	}
	return true
}

func HandleSelectPokemon(selectedPokemonMsg *SelectPokemonMessage, issuer *TrainerBattleStatus,
	issuerChan chan ws.Serializable) bool {
	selectedPokemonId := selectedPokemonMsg.PokemonId
	pokemon, ok := issuer.TrainerPokemons[selectedPokemonId]
	if !ok {
		issuerChan <- ws.ErrorMessage{
			Info:  fmt.Sprintf(ErrorInvalidPokemonSelected.Error()),
			Fatal: false,
		}
		return false
	}

	if pokemon.HP <= 0 {
		issuerChan <- ws.ErrorMessage{
			Info:  fmt.Sprintf(ErrorPokemonNoHP.Error()),
			Fatal: false,
		}
		return false
	}
	issuer.SelectedPokemon = pokemon
	log.Info("Changed selected pokemon")
	UpdateTrainerPokemon(selectedPokemonMsg.TrackedMessage, *issuer.SelectedPokemon, issuerChan, true)
	return true
}

func HandleDefendMove(issuer *TrainerBattleStatus, issuerChan chan ws.Serializable, cooldownDuration time.Duration) {
	// if the pokemon is dead, player must select a new pokemon
	if issuer.SelectedPokemon.HP == 0 {
		issuerChan <- ws.ErrorMessage{
			Info:  fmt.Sprintf(ErrorPokemonNoHP.Error()),
			Fatal: false,
		}
		return
	}

	// if player has moved recently and is in Cooldown, discard move
	if issuer.Cooldown {
		issuerChan <- ws.ErrorMessage{
			Info:  fmt.Sprintf(ErrorCooldown.Error()),
			Fatal: false,
		}
		return
	}
	issuer.CdTimer.Reset(cooldownDuration)
	issuer.Cooldown = true

	// process Defending move: update both players and setup a Cooldown
	issuer.Defending = true
	issuerChan <- StatusMessage{
		Message: StatusDefended,
	}
}

func HandleAttackMove(issuer *TrainerBattleStatus, issuerChan chan ws.Serializable, defending bool,
	otherPokemon *pokemons.Pokemon, cooldownDuration time.Duration) bool {
	if issuer.SelectedPokemon.HP == 0 {
		issuerChan <- ws.ErrorMessage{
			Info:  fmt.Sprintf(ErrorPokemonNoHP.Error()),
			Fatal: false,
		}
		return false
	}

	// if player has moved recently and is in Cooldown, discard move
	if issuer.Cooldown {
		issuerChan <- ws.ErrorMessage{
			Info:  fmt.Sprintf(ErrorCooldown.Error()),
			Fatal: false,
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

func UpdateTrainerPokemon(trackedMsg ws.TrackedMessage, pokemon pokemons.Pokemon, channel chan ws.Serializable, owner bool) {
	channel <- UpdatePokemonMessage{
		Owner:          owner,
		Pokemon:        pokemon,
		TrackedMessage: trackedMsg,
	}
}
