package tokens

import (
	"crypto/md5"
	"encoding/json"
	"fmt"
	"github.com/NOVAPokemon/utils"
	"github.com/dgrijalva/jwt-go"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"net/http"
	"strings"
	"time"
)

const (
	JWTDuration = 30 * time.Minute

	AuthTokenHeaderName     = "Auth_token"
	StatsTokenHeaderName    = "Stats_token"
	PokemonsTokenHeaderName = "Pokemon_token"
	ItemsTokenHeaderName    = "Items_token"
)

var (
	authJWTKey = []byte("authJWTKey")
)

func ExtractAndVerifyAuthToken(headers http.Header) (*AuthToken, error) {

	tknStr := headers.Get(AuthTokenHeaderName)
	authToken := &AuthToken{}
	jwtToken, err := jwt.ParseWithClaims(tknStr, authToken, func(token *jwt.Token) (interface{}, error) {
		return authJWTKey, nil
	})

	if err != nil {
		return nil, err
	}

	if !jwtToken.Valid {
		return nil, errors.New("Invalid Token")
	}

	return authToken, err
}

func ExtractAndVerifyTrainerStatsToken(headers http.Header) (*TrainerStatsToken, error) {

	tknStr := headers.Get(StatsTokenHeaderName)
	statsTkn := &TrainerStatsToken{}
	jwtToken, err := jwt.ParseWithClaims(tknStr, statsTkn, func(token *jwt.Token) (interface{}, error) {
		return authJWTKey, nil
	})

	if err != nil {
		return nil, err
	}

	if !jwtToken.Valid {
		return nil, errors.New("Invalid Token")
	}

	return statsTkn, err
}

func ExtractAndVerifyPokemonTokens(headers http.Header) (map[string]PokemonToken, error) {

	var pokemonTkns = make(map[string]PokemonToken, len(headers))

	for name, v := range headers {
		logrus.Infof(fmt.Sprintf("Header value: %s", name))
		fmt.Println(strings.Contains(name, PokemonsTokenHeaderName))
		if strings.Contains(name, PokemonsTokenHeaderName) {
			tknStr := strings.Join(v, "")
			pokemonTkn := &PokemonToken{}
			_, err := jwt.ParseWithClaims(tknStr, pokemonTkn, func(token *jwt.Token) (interface{}, error) {
				return authJWTKey, nil
			})
			if err != nil {
				return nil, err
			}
			pokemonTkns[name] = *pokemonTkn
		}
	}

	return pokemonTkns, nil
}

func ExtractAndVerifyItemsToken(headers http.Header) (*ItemsToken, error) {

	tknStr := headers.Get(ItemsTokenHeaderName)
	itemsToken := &ItemsToken{}
	jwtToken, err := jwt.ParseWithClaims(tknStr, itemsToken, func(token *jwt.Token) (interface{}, error) {
		return authJWTKey, nil
	})

	if err != nil {
		return nil, err
	}

	if !jwtToken.Valid {
		return nil, errors.New("Invalid Token")
	}

	return itemsToken, err
}

func ExtractItemsToken(itemsToken string) (*ItemsToken, error) {
	claims := ItemsToken{}
	_, _, err := new(jwt.Parser).ParseUnverified(itemsToken, &claims)
	if err != nil {
		log.Error("error parsing items token, ", err)
		return nil, err
	}

	return &claims, err
}

func ExtractPokemonToken(pokemonsToken string) (*PokemonToken, error) {
	claims := PokemonToken{}
	_, _, err := new(jwt.Parser).ParseUnverified(pokemonsToken, &claims)
	if err != nil {
		log.Error("error parsing pokemon token, ", err)
		return nil, err
	}

	return &claims, err
}

func ExtractStatsToken(statsToken string) (*TrainerStatsToken, error) {
	claims := TrainerStatsToken{}
	_, _, err := new(jwt.Parser).ParseUnverified(statsToken, &claims)
	if err != nil {
		log.Error("error parsing stats token, ", err)
		return nil, err
	}

	return &claims, err
}

func AddAuthToken(username string, headers http.Header) {
	expirationTime := time.Now().Add(JWTDuration)
	authToken := &AuthToken{
		Username:       username,
		StandardClaims: jwt.StandardClaims{ExpiresAt: expirationTime.Unix()},
	}
	setTokenInHeader(AuthTokenHeaderName, authToken, headers)
}

func AddPokemonsTokens(pokemons map[string]utils.Pokemon, headers http.Header) {
	expirationTime := time.Now().Add(JWTDuration)
	for k, v := range pokemons {
		pokemonToken := &PokemonToken{
			Pokemon:        v,
			PokemonHash:    generateHash(v),
			StandardClaims: jwt.StandardClaims{ExpiresAt: expirationTime.Unix()},
		}
		log.Infof(fmt.Sprintf("Putting token: %s-%s", PokemonsTokenHeaderName, k))
		setTokenInHeader(fmt.Sprintf("%s-%s", PokemonsTokenHeaderName, k), pokemonToken, headers)
	}

}

func AddItemsToken(items map[string]utils.Item, headers http.Header) {
	expirationTime := time.Now().Add(JWTDuration)
	trainerItemsToken := &ItemsToken{
		Items:          items,
		ItemsHash:      generateHash(items),
		StandardClaims: jwt.StandardClaims{ExpiresAt: expirationTime.Unix()},
	}
	setTokenInHeader(ItemsTokenHeaderName, trainerItemsToken, headers)
}

func AddTrainerStatsToken(stats utils.TrainerStats, headers http.Header) {
	expirationTime := time.Now().Add(JWTDuration)
	trainerStatsToken := &TrainerStatsToken{
		TrainerStats:   stats,
		TrainerHash:    generateHash(&stats),
		StandardClaims: jwt.StandardClaims{ExpiresAt: expirationTime.Unix()},
	}
	setTokenInHeader(StatsTokenHeaderName, trainerStatsToken, headers)
}

func setTokenInHeader(headerName string, token interface{ jwt.Claims }, headers http.Header) {
	jwtToken := jwt.NewWithClaims(jwt.SigningMethodHS256, token)
	tokenString, err := jwtToken.SignedString(authJWTKey)
	if err != nil {
		panic(err)
	}
	headers.Set(headerName, tokenString)
}

func generateHash(toHash interface{}) []byte {
	marshaled, _ := json.Marshal(toHash)
	hash := md5.Sum(marshaled)
	return hash[:]
}
