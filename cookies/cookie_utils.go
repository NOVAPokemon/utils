package cookies

import (
	"crypto/md5"
	"encoding/json"
	"fmt"
	"github.com/NOVAPokemon/utils"
	"github.com/dgrijalva/jwt-go"
	"github.com/pkg/errors"
	"net/http"
	"strings"
	"time"
)

var ErrInvalidToken = errors.New("Invalid Token")

const (
	JWTDuration = 30 * time.Minute

	AuthTokenCookieName     = "auth_token"
	StatsTokenCookieName    = "stats_token"
	PokemonsTokenCookieName = "pokemon_token"
	ItemsTokenCookieName    = "items_token"
)

var (
	authJWTKey = []byte("authJWTKey")
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

func ExtractTrainerStatsToken(r *http.Request) (statsTkn *TrainerStatsToken, err error) {

	c, err := r.Cookie(StatsTokenCookieName)

	if err != nil {
		return nil, err
	}

	tknStr := c.Value
	statsTkn = &TrainerStatsToken{}
	_, err = jwt.ParseWithClaims(tknStr, statsTkn, func(token *jwt.Token) (interface{}, error) {
		return authJWTKey, nil
	})
	if err != nil {
		return nil, err
	}

	return statsTkn, nil
}

func ExtractPokemonTokens(r *http.Request) (pokemonsTkns []PokemonToken) {
	for _, v := range r.Cookies() {
		if strings.Contains(v.Name, PokemonsTokenCookieName) {
			tknStr := v.Value
			pokemonTkn := &PokemonToken{}
			_, err := jwt.ParseWithClaims(tknStr, pokemonTkn, func(token *jwt.Token) (interface{}, error) {
				return authJWTKey, nil
			})
			if err != nil {
				continue
			}
			pokemonsTkns = append(pokemonsTkns, *pokemonTkn)
		}
	}

	return pokemonsTkns
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

func SetPokemonsCookies(pokemons map[string]utils.Pokemon, w http.ResponseWriter) {
	expirationTime := time.Now().Add(JWTDuration)

	for k, v := range pokemons {
		pokemonToken := &PokemonToken{
			Pokemon:        v,
			PokemonHash:    generateHash(v),
			StandardClaims: jwt.StandardClaims{ExpiresAt: expirationTime.Unix()},
		}

		SetTokenAsCookie(fmt.Sprintf("%s-%s", PokemonsTokenCookieName, k), pokemonToken, w)
	}

}

func SetItemsCookie(items map[string]utils.Item, w http.ResponseWriter) {
	expirationTime := time.Now().Add(JWTDuration)
	trainerItemsToken := &ItemsToken{
		Items:          items,
		ItemsHash:      generateHash(items),
		StandardClaims: jwt.StandardClaims{ExpiresAt: expirationTime.Unix()},
	}

	SetTokenAsCookie(ItemsTokenCookieName, trainerItemsToken, w)

}

func SetTrainerStatsCookie(stats utils.TrainerStats, w http.ResponseWriter) {

	expirationTime := time.Now().Add(JWTDuration)
	trainerStatsToken := &TrainerStatsToken{
		TrainerStats:   stats,
		TrainerHash:    generateHash(&stats),
		StandardClaims: jwt.StandardClaims{ExpiresAt: expirationTime.Unix()},
	}

	SetTokenAsCookie(StatsTokenCookieName, trainerStatsToken, w)

}

func SetTokenAsCookie(name string, token interface{ jwt.Claims }, w http.ResponseWriter) {
	jwtToken := jwt.NewWithClaims(jwt.SigningMethodHS256, token)
	tokenString, err := jwtToken.SignedString(authJWTKey)

	if err != nil {
		panic(err)
	}

	http.SetCookie(w,
		&http.Cookie{
			Name:    name,
			Value:   tokenString,
			Path:    "/",
			Domain:  utils.Host,
			Expires: time.Now().Add(JWTDuration),
		})
}

func generateHash(toHash interface{}) []byte {
	marshaled, _ := json.Marshal(toHash)
	hash := md5.Sum(marshaled)
	return hash[:]
}
