package tokens

import (
	"fmt"
	"github.com/NOVAPokemon/utils"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"net/http"
	"os"
	"testing"
)

const username = "test_trainer"

var id1 = primitive.NewObjectID()
var id2 = primitive.NewObjectID()
var id3 = primitive.NewObjectID()

var pokemons = map[string]utils.Pokemon{
	id1.Hex(): utils.Pokemon{Id: id1},
	id2.Hex(): utils.Pokemon{Id: id2},
	id3.Hex(): utils.Pokemon{Id: id3},
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

	AddPokemonsTokens(pokemons, header)
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
		correpondingPokemon := pokemons[token.Pokemon.Id.Hex()]
		logrus.Infof("%+v-%+v", correpondingPokemon, token.Pokemon)
		assert.Equal(t, correpondingPokemon, token.Pokemon)
	}
}
