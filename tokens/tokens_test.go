package tokens

import (
	"fmt"
	"github.com/NOVAPokemon/utils"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"net/http"
	"os"
	"testing"
)

const username = "test_trainer"

var pokemons = map[string]utils.Pokemon{
	"pokemon1": utils.Pokemon{},
	"pokemon2": utils.Pokemon{},
	"pokemon3": utils.Pokemon{},
}

var items = map[string]utils.Item{
	"item1": utils.Item{},
	"item2": utils.Item{},
	"item3": utils.Item{},
}

func TestMain(m *testing.M) {
	res := m.Run()
	os.Exit(res)
}

func TestAuthToken(t *testing.T) {
	header := http.Header{}
	AddAuthToken(username, header)
	token, err := ExtractAndVerifyAuthToken(header)

	if err != nil {
		t.Error(err)
		t.FailNow()
	}

	assert.Equal(t, token.Username, username)
}

func TestItemsToken(t *testing.T) {
	header := http.Header{}
	AddItemsToken(items, header)
	token, err := ExtractAndVerifyItemsToken(header)

	if err != nil {
		t.Error(err)
		t.FailNow()
	}

	for k, item := range token.Items {
		correpondingItem := items[k]
		assert.Equal(t, correpondingItem, item)
	}
}

func TestPokemonToken(t *testing.T) {
	header := http.Header{}

	fmt.Println("here")

	AddPokemonsTokens(pokemons, header)
	tokens, err := ExtractAndVerifyPokemonTokens(header)

	fmt.Println("here2")

	if err != nil {
		t.Error(err)
		t.FailNow()
	}

	if len(tokens) == 0 {
		t.Error("pokemon tokens empty")
		t.FailNow()
	}

	for k, token := range tokens {
		fmt.Println("here4")
		fmt.Println(k, token)
		correpondingPokemon := pokemons[k]
		logrus.Infof("%+v-%+v", correpondingPokemon, token.Pokemon)
		assert.Equal(t, correpondingPokemon, token.Pokemon)
	}
}
