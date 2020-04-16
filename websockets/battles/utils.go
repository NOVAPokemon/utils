package battles

import (
	"encoding/json"
	"fmt"
	"github.com/NOVAPokemon/utils/pokemons"
	"github.com/NOVAPokemon/utils/websockets"
	log "github.com/sirupsen/logrus"
)

func HandleUseItem(message *websockets.Message, issuer *TrainerBattleStatus, issuerChan chan *string) error {

	if len(message.MsgArgs) < 1 {
		errMsg := websockets.Message{MsgType: ERROR, MsgArgs: []string{ErrNoItemSelected.Error()}}
		websockets.SendMessage(errMsg, issuerChan)
		return ErrInvalidItemSelected
	}

	if issuer.Cooldown {
		errMsg := websockets.Message{MsgType: ERROR, MsgArgs: []string{ErrCooldown.Error()}}
		websockets.SendMessage(errMsg, issuerChan)
		return ErrCooldown
	}

	selectedItem := message.MsgArgs[0]
	item, ok := issuer.TrainerItems[selectedItem]

	if !ok {
		errMsg := websockets.Message{MsgType: ERROR, MsgArgs: []string{ErrInvalidItemSelected.Error()}}
		websockets.SendMessage(errMsg, issuerChan)
		return ErrInvalidItemSelected
	}

	if !item.Effect.Appliable {
		errMsg := websockets.Message{MsgType: ERROR, MsgArgs: []string{ErrItemNotAppliable.Error()}}
		websockets.SendMessage(errMsg, issuerChan)
		return ErrItemNotAppliable
	}

	err := item.Apply(issuer.SelectedPokemon)

	if err != nil {
		errMsg := websockets.Message{MsgType: ERROR, MsgArgs: []string{err.Error()}}
		websockets.SendMessage(errMsg, issuerChan)
		return err
	}

	issuer.CdTimer.Reset(DefaultCooldown)
	issuer.Cooldown = true

	issuer.UsedItems[item.Id.Hex()] = item
	delete(issuer.TrainerItems, item.Id.Hex())
	UpdateTrainerPokemon(*issuer.SelectedPokemon, issuerChan)
	msg := websockets.Message{MsgType: REMOVE_ITEM, MsgArgs: []string{string(item.Id.Hex())}}
	websockets.SendMessage(msg, issuerChan)
	return nil
}

func HandleSelectPokemon(msgStr *string, issuer *TrainerBattleStatus, issuerChan chan *string) error {

	message, err := websockets.ParseMessage(msgStr)

	if err != nil {
		errMsg := websockets.Message{MsgType: ERROR, MsgArgs: []string{ErrPokemonSelectionPhase.Error()}}
		websockets.SendMessage(errMsg, issuerChan)
		return err
	}

	if message.MsgType != SELECT_POKEMON {
		errMsg := websockets.Message{MsgType: ERROR, MsgArgs: []string{ErrPokemonSelectionPhase.Error()}}
		websockets.SendMessage(errMsg, issuerChan)
		return err
	}

	if len(message.MsgArgs) < 1 {
		errMsg := websockets.Message{MsgType: ERROR, MsgArgs: []string{ErrNoPokemonSelected.Error()}}
		websockets.SendMessage(errMsg, issuerChan)
		return err
	}

	selectedPokemon := message.MsgArgs[0]
	pokemon, ok := issuer.TrainerPokemons[selectedPokemon]

	if !ok {
		errMsg := websockets.Message{MsgType: ERROR, MsgArgs: []string{ErrInvalidPokemonSelected.Error()}}
		websockets.SendMessage(errMsg, issuerChan)
		return err
	}
	if pokemon.HP <= 0 {
		// pokemon is dead
		msg := websockets.Message{MsgType: ERROR, MsgArgs: []string{fmt.Sprintf(ErrPokemonNoHP.Error())}}
		websockets.SendMessage(msg, issuerChan)
	}
	issuer.SelectedPokemon = pokemon
	UpdateTrainerPokemon(*issuer.SelectedPokemon, issuerChan)
	return nil
}

func HandleDefendMove(issuer *TrainerBattleStatus, issuerChan chan *string) error {

	// if the pokemon is dead, player must select a new pokemon
	if issuer.SelectedPokemon.HP == 0 {
		errMsg := websockets.Message{MsgType: ERROR, MsgArgs: []string{ErrPokemonNoHP.Error()}}
		websockets.SendMessage(errMsg, issuerChan)
		return ErrPokemonNoHP
	}

	// if player has moved recently and is in Cooldown, discard move
	if issuer.Cooldown {
		errMsg := websockets.Message{MsgType: ERROR, MsgArgs: []string{ErrCooldown.Error()}}
		websockets.SendMessage(errMsg, issuerChan)
		return ErrCooldown
	}
	issuer.CdTimer.Reset(DefaultCooldown)
	issuer.Cooldown = true

	// process Defending move: update both players and setup a Cooldown
	issuer.Defending = true
	msg := websockets.Message{MsgType: DEFEND_SUCCESS, MsgArgs: []string{DefaultCooldown.String()}}
	websockets.SendMessage(msg, issuerChan)
	return nil
}

func HandleAttackMove(issuer *TrainerBattleStatus, issuerChan chan *string, defending bool, otherPokemon *pokemons.Pokemon) (bool, error) {

	if issuer.SelectedPokemon.HP == 0 {
		errMsg := websockets.Message{MsgType: ERROR, MsgArgs: []string{ErrPokemonNoHP.Error()}}
		websockets.SendMessage(errMsg, issuerChan)
		return false, ErrPokemonNoHP
	}

	// if player has moved recently and is in Cooldown, discard move
	if issuer.Cooldown {
		errMsg := websockets.Message{MsgType: ERROR, MsgArgs: []string{ErrCooldown.Error()}}
		websockets.SendMessage(errMsg, issuerChan)
		return false, ErrPokemonNoHP
	}

	issuer.CdTimer.Reset(DefaultCooldown)
	issuer.Cooldown = true

	if defending {
		msg := websockets.Message{MsgType: STATUS, MsgArgs: []string{StatusOpponentDefended}}
		websockets.SendMessage(msg, issuerChan)
		return false, nil
	} else {
		otherPokemon.HP -= issuer.SelectedPokemon.Damage
		if otherPokemon.HP < 0 {
			otherPokemon.HP = 0
		}
		return true, nil
	}

}

func UpdateTrainerPokemon(pokemon pokemons.Pokemon, ownerChan chan *string) {

	toSend, err := json.Marshal(pokemon)

	if err != nil {
		log.Error(err)
		return
	}

	msg := websockets.Message{MsgType: UPDATE_PLAYER_POKEMON, MsgArgs: []string{string(toSend)}}
	websockets.SendMessage(msg, ownerChan)
}

func UpdateAdversaryOfPokemonChanges(pokemon pokemons.Pokemon, adversaryChan chan *string) {

	toSend, err := json.Marshal(pokemon)

	if err != nil {
		log.Error(err)
		return
	}

	msg := websockets.Message{MsgType: UPDATE_ADVERSARY_POKEMON, MsgArgs: []string{string(toSend)}}
	websockets.SendMessage(msg, adversaryChan)
}
