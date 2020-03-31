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

	AuthTokenCookieName     = "auth_token"
	StatsTokenCookieName    = "stats_token"
	PokemonsTokenCookieName = "pokemons_token"
	ItemsTokenCookieName    = "items_token"
)

var (
	authJWTKey = []byte("authJWTKey") // TODO change
)

func ExtractAndVerifyAuthToken(w *http.ResponseWriter, r *http.Request, caller string) (authToken *AuthToken, err error) {
	c, err := r.Cookie(AuthTokenCookieName)

	if err != nil {
		utils.HandleCookieError(w, caller, err)
		return nil, err
	}

	tknStr := c.Value
	authToken = &AuthToken{}
	tkn, err := jwt.ParseWithClaims(tknStr, authToken, func(token *jwt.Token) (interface{}, error) {
		return authJWTKey, nil
	})

	if err != nil {
		utils.HandleJWTVerifyingError(w, caller, err)
		return nil, err
	}

	if !tkn.Valid || time.Unix(authToken.StandardClaims.ExpiresAt, 0).Unix() < time.Now().Unix() {
		(*w).WriteHeader(http.StatusUnauthorized)
		return nil, err
	}

	return authToken, nil
}

func ExtractTrainerStatsToken(r *http.Request) (trainerStats *TrainerStatsToken, err error) {

	c, err := r.Cookie(StatsTokenCookieName)

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
	c, err := r.Cookie(PokemonsTokenCookieName)

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

func ExtractItemsToken(r *http.Request) (itemsTkn *ItemsToken, err error) {
	c, err := r.Cookie(ItemsTokenCookieName)

	if err != nil {
		return nil, err
	}

	tknStr := c.Value
	itemsTkn = &ItemsToken{}
	_, err = jwt.ParseWithClaims(tknStr, itemsTkn, func(token *jwt.Token) (interface{}, error) {
		return authJWTKey, nil
	})

	if err != nil {
		return nil, err
	}

	return itemsTkn, nil
}

func SetAuthToken(username, caller string, w *http.ResponseWriter) error {
	expirationTime := time.Now().Add(JWTDuration)
	claims := &AuthToken{
		Username:       username,
		StandardClaims: jwt.StandardClaims{ExpiresAt: expirationTime.Unix()},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(authJWTKey)

	if err != nil {
		utils.HandleJWTSigningError(w, caller, err)
		return err
	}

	http.SetCookie(*w,
		&http.Cookie{
			Name:    AuthTokenCookieName,
			Value:   tokenString,
			Path:    "/",
			Domain:  utils.Host,
			Expires: time.Now().Add(JWTDuration),
		})

	return nil
}

func SetPokemonsCookie(pokemons map[string]utils.Pokemon, w http.ResponseWriter) {
	expirationTime := time.Now().Add(JWTDuration)
	trainerStatsToken := &PokemonsToken{
		Pokemons:       pokemons,
		PokemonHashes:  generatePokemonHashes(pokemons),
		StandardClaims: jwt.StandardClaims{ExpiresAt: expirationTime.Unix()},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, trainerStatsToken)
	tokenString, err := token.SignedString(authJWTKey)

	if err != nil {
		panic(err)
	}

	http.SetCookie(w,
		&http.Cookie{
			Name:    PokemonsTokenCookieName,
			Value:   tokenString,
			Path:    "/",
			Domain:  utils.Host,
			Expires: time.Now().Add(JWTDuration),
		})
}

func SetItemsCookie(items map[string]utils.Item, w http.ResponseWriter) {
	expirationTime := time.Now().Add(JWTDuration)
	trainerItemsToken := &ItemsToken{
		Items:          items,
		ItemsHash:      generateItemsHash(items),
		StandardClaims: jwt.StandardClaims{ExpiresAt: expirationTime.Unix()},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, trainerItemsToken)
	tokenString, err := token.SignedString(authJWTKey)

	if err != nil {
		panic(err)
	}

	http.SetCookie(w,
		&http.Cookie{
			Name:    ItemsTokenCookieName,
			Value:   tokenString,
			Path:    "/",
			Domain:  utils.Host,
			Expires: time.Now().Add(JWTDuration),
		})
}

func SetTrainerStatsCookie(stats utils.TrainerStats, w http.ResponseWriter) {
	expirationTime := time.Now().Add(JWTDuration)
	trainerStatsToken := &TrainerStatsToken{
		TrainerStats:   stats,
		TrainerHash:    generateTrainerStatsHash(stats),
		StandardClaims: jwt.StandardClaims{ExpiresAt: expirationTime.Unix()},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, trainerStatsToken)
	tokenString, err := token.SignedString(authJWTKey)

	if err != nil {
		panic(err)
	}

	http.SetCookie(w,
		&http.Cookie{
			Name:    StatsTokenCookieName,
			Value:   tokenString,
			Path:    "/",
			Domain:  utils.Host,
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
