package cookies

import (
	"crypto/md5"
	"encoding/json"
	"github.com/NOVAPokemon/utils"
	"github.com/dgrijalva/jwt-go"
	"github.com/pkg/errors"
	"net/http"
	"time"
)

var ErrInvalidToken = errors.New("Invalid Token")

const (
	JWTDuration = 30 * time.Minute

	authTokenCookieName     = "auth_token"
	statsTokenCookieName    = "stats_token"
	pokemonsTokenCookieName = "pokemons_token"
	itemsTokenCookieName    = "items_token"
)

var (
	authJWTKey = []byte("authJWTKey") // TODO change
)

func ExtractAndVerifyAuthToken(w *http.ResponseWriter, r *http.Request, caller string) (authToken *AuthToken, err error) {
	c, err := r.Cookie(authTokenCookieName)

	if err != nil {
		utils.HandleCookieError(w, caller, err)
		return nil, err
	}

	tknStr := c.Value
	authToken = &AuthToken{}
	tkn, err := jwt.ParseWithClaims(tknStr, authToken.Claims, func(token *jwt.Token) (interface{}, error) {
		return authJWTKey, nil
	})

	if err != nil {
		utils.HandleJWTVerifyingError(w, caller, err)
		return nil, err
	}

	if !tkn.Valid && time.Unix(authToken.Claims.ExpiresAt, 0).Unix() < time.Now().Unix() {
		(*w).WriteHeader(http.StatusUnauthorized)
		return nil, err
	}

	return authToken, nil
}

func ExtractTrainerStatsToken(r *http.Request) (trainerStats *TrainerStatsToken, err error) {

	c, err := r.Cookie(statsTokenCookieName)

	if err != nil {
		return nil, err
	}

	tknStr := c.Value
	trainerStats = &TrainerStatsToken{}
	err = json.Unmarshal([]byte(tknStr), trainerStats)

	if err != nil {
		return nil, err
	}

	return trainerStats, nil
}

func ExtractPokemonsToken(r *http.Request) (pokemons *PokemonsToken, err error) {
	c, err := r.Cookie(pokemonsTokenCookieName)

	if err != nil {
		return nil, err
	}

	tknStr := c.Value

	pokemons = &PokemonsToken{}
	err = json.Unmarshal([]byte(tknStr), pokemons)

	if err != nil {
		return nil, err
	}

	return pokemons, nil
}

func ExtractItemsToken(r *http.Request) (items *ItemsToken, err error) {
	c, err := r.Cookie(itemsTokenCookieName)

	if err != nil {
		return nil, err
	}

	tknStr := c.Value
	items = &ItemsToken{}
	err = json.Unmarshal([]byte(tknStr), items)

	if err != nil {
		return nil, err
	}

	return items, nil
}

func SetPokemonsCookie(pokemons map[string]utils.Pokemon, w http.ResponseWriter, key []byte) {

	expirationTime := time.Now().Add(JWTDuration)
	trainerStatsToken := &PokemonsToken{
		Pokemons:      pokemons,
		PokemonHashes: generatePokemonHashes(pokemons),
		Claims:        jwt.StandardClaims{ExpiresAt: expirationTime.Unix()},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, trainerStatsToken.Claims)
	tokenString, err := token.SignedString(key)

	if err != nil {
		panic(err)
	}

	http.SetCookie(w,
		&http.Cookie{
			Name:    itemsTokenCookieName,
			Value:   tokenString,
			Expires: time.Now().Add(JWTDuration),
		})
}

func SetItemsCookie(items map[string]utils.Item, w http.ResponseWriter, key []byte) {

	expirationTime := time.Now().Add(JWTDuration)
	trainerItemsToken := &ItemsToken{
		Items:     items,
		ItemsHash: generateItemsHash(items),
		Claims:    jwt.StandardClaims{ExpiresAt: expirationTime.Unix()},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, trainerItemsToken.Claims)
	tokenString, err := token.SignedString(key)

	if err != nil {
		panic(err)
	}

	http.SetCookie(w,
		&http.Cookie{
			Name:    itemsTokenCookieName,
			Value:   tokenString,
			Expires: time.Now().Add(JWTDuration),
		})
}

func SetTrainerStatsCookie(stats utils.TrainerStats, w http.ResponseWriter, key []byte) {

	expirationTime := time.Now().Add(JWTDuration)
	trainerStatsToken := &TrainerStatsToken{
		TrainerStats: stats,
		TrainerHash:  generateTrainerStatsHash(stats),
		Claims:       jwt.StandardClaims{ExpiresAt: expirationTime.Unix()},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, trainerStatsToken.Claims)
	tokenString, err := token.SignedString(key)

	if err != nil {
		panic(err)
	}

	http.SetCookie(w,
		&http.Cookie{
			Name:    statsTokenCookieName,
			Value:   tokenString,
			Expires: time.Now().Add(JWTDuration),
		})
}

func generatePokemonHashes(pokemons map[string]utils.Pokemon) map[string][]byte {
	toReturn := make(map[string][]byte, len(pokemons))
	for pokemonId, pokemon := range pokemons {
		marshaled, _ := json.Marshal(pokemon)
		hash := md5.Sum(marshaled)
		toReturn[pokemonId] = hash[:]
	}
	return toReturn
}

func generateTrainerStatsHash(stats utils.TrainerStats) []byte {
	marshaled, _ := json.Marshal(stats)
	hash := md5.Sum(marshaled)
	return hash[:]
}

func generateItemsHash(items map[string]utils.Item) []byte {
	marshaled, _ := json.Marshal(items)
	hash := md5.Sum(marshaled)
	return hash[:]
}
