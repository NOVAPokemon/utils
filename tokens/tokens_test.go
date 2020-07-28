package tokens

import (
	"fmt"
	"net/http"
	"os"
	"testing"

	"github.com/NOVAPokemon/utils/items"
	"github.com/NOVAPokemon/utils/pokemons"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

const username = "test_trainer"

var id1 = primitive.NewObjectID()
var id2 = primitive.NewObjectID()
var id3 = primitive.NewObjectID()

var pokemonsTest = map[string]pokemons.Pokemon{
	id1.Hex(): {Id: id1.Hex()},
	id2.Hex(): {Id: id2.Hex()},
	id3.Hex(): {Id: id3.Hex()},
}

var itemsToTest = map[string]items.Item{
	"item1": {},
	"item2": {},
	"item3": {},
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
	AddItemsToken(itemsToTest, header)
	token, err := ExtractAndVerifyItemsToken(header)

	if err != nil {
		t.Error(err)
		t.FailNow()
	}

	for k, item := range token.Items {
		correpondingItem := itemsToTest[k]
		assert.Equal(t, correpondingItem, item)
	}
}

func TestPokemonToken(t *testing.T) {
	header := http.Header{}

	AddPokemonsTokens(pokemonsTest, header)
	tokens, err := ExtractAndVerifyPokemonTokens(header)

	if err != nil {
		t.Error(err)
		t.FailNow()
	}

	if len(tokens) == 0 {
		t.Error("pokemon tokens empty")
		t.FailNow()
	}

	for k, token := range tokens {
		fmt.Println(k, token)
		correpondingPokemon := pokemonsTest[token.Pokemon.Id]
		logrus.Infof("%+v-%+v", correpondingPokemon, token.Pokemon)
		assert.Equal(t, correpondingPokemon, token.Pokemon)
	}
}
