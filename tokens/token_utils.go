package tokens

import (
	"crypto/md5"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/NOVAPokemon/utils"
	"github.com/NOVAPokemon/utils/items"
	"github.com/NOVAPokemon/utils/pokemons"
	"github.com/dgrijalva/jwt-go"
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
		err = wrapExtractVerifyAuthTokenError(wrapParsingTokenError(err))
		return nil, err
	}

	if !jwtToken.Valid {
		err = wrapExtractVerifyAuthTokenError(ErrorInvalidAuthToken)
		return nil, err
	}

	return authToken, nil
}

func ExtractAndVerifyTrainerStatsToken(headers http.Header) (*TrainerStatsToken, error) {
	tknStr := headers.Get(StatsTokenHeaderName)
	statsTkn := &TrainerStatsToken{}

	jwtToken, err := jwt.ParseWithClaims(tknStr, statsTkn, func(token *jwt.Token) (interface{}, error) {
		return authJWTKey, nil
	})
	if err != nil {
		err = wrapExtractVerifyStatsTokenError(wrapParsingTokenError(err))
		return nil, err
	}

	if !jwtToken.Valid {
		err = wrapExtractVerifyStatsTokenError(ErrorInvalidStatsToken)
		return nil, err
	}

	return statsTkn, nil
}

func ExtractAndVerifyPokemonTokens(headers http.Header) ([]*PokemonToken, error) {
	tkns, ok := headers[PokemonsTokenHeaderName]
	if !ok {
		err := wrapExtractVerifyPokemonTokensError(ErrorNoPokemonTokens)
		return nil, err
	}

	var pokemonTkns = make([]*PokemonToken, len(tkns))
	i := 0
	for ; i < len(tkns); i++ {
		if len(tkns[i]) == 0 {
			continue
		}

		pokemonTkn := PokemonToken{}
		_, err := jwt.ParseWithClaims(strings.TrimSpace(tkns[i]), &pokemonTkn,
			func(token *jwt.Token) (interface{}, error) {
				return authJWTKey, nil
			})
		if err != nil {
			err = wrapExtractVerifyPokemonTokensError(wrapParsingTokenError(err))
			return nil, err
		}

		pokemonTkns[i] = &pokemonTkn
	}

	return pokemonTkns[:i], nil
}

func ExtractAndVerifyItemsToken(headers http.Header) (*ItemsToken, error) {
	tknStr := headers.Get(ItemsTokenHeaderName)
	itemsToken := &ItemsToken{}

	jwtToken, err := jwt.ParseWithClaims(tknStr, itemsToken, func(token *jwt.Token) (interface{}, error) {
		return authJWTKey, nil
	})
	if err != nil {
		err = wrapExtractVerifyItemsTokenError(wrapParsingTokenError(err))
		return nil, err
	}

	if !jwtToken.Valid {
		err = wrapExtractVerifyItemsTokenError(ErrorInvalidItemsToken)
		return nil, err
	}

	return itemsToken, nil
}

func ExtractStatsToken(statsToken string) (*TrainerStatsToken, error) {
	claims := TrainerStatsToken{}

	_, _, err := new(jwt.Parser).ParseUnverified(statsToken, &claims)
	if err != nil {
		err = wrapExtractStatsTokenError(wrapParsingTokenError(err))
		return nil, err
	}

	return &claims, nil
}

func ExtractPokemonToken(pokemonsToken string) (*PokemonToken, error) {
	claims := PokemonToken{}

	_, _, err := new(jwt.Parser).ParseUnverified(pokemonsToken, &claims)
	if err != nil {
		err = wrapExtractPokemonTokenError(wrapParsingTokenError(err))
		return nil, err
	}

	return &claims, nil
}

func ExtractItemsToken(itemsToken string) (*ItemsToken, error) {
	claims := ItemsToken{}
	_, _, err := new(jwt.Parser).ParseUnverified(itemsToken, &claims)
	if err != nil {
		err = wrapExtractItemsTokenError(wrapParsingTokenError(err))
		return nil, err
	}

	return &claims, nil
}

func AddAuthToken(username string, headers http.Header) {
	expirationTime := time.Now().Add(JWTDuration)
	authToken := &AuthToken{
		Username:       username,
		StandardClaims: jwt.StandardClaims{ExpiresAt: expirationTime.Unix()},
	}

	setTokenInHeader(AuthTokenHeaderName, authToken, headers)
}

func AddTrainerStatsToken(stats utils.TrainerStats, headers http.Header) {
	expirationTime := time.Now().Add(JWTDuration)
	trainerStatsToken := &TrainerStatsToken{
		TrainerStats:   stats,
		TrainerHash:    generateHash(stats),
		StandardClaims: jwt.StandardClaims{ExpiresAt: expirationTime.Unix()},
	}
	setTokenInHeader(StatsTokenHeaderName, trainerStatsToken, headers)
}

func AddPokemonsTokens(pokemons map[string]pokemons.Pokemon, headers http.Header) {
	expirationTime := time.Now().Add(JWTDuration)
	for _, v := range pokemons {
		pokemonToken := &PokemonToken{
			Pokemon:        v,
			PokemonHash:    generateHash(v),
			StandardClaims: jwt.StandardClaims{ExpiresAt: expirationTime.Unix()},
		}
		addTokenToHeader(PokemonsTokenHeaderName, pokemonToken, headers)
	}
}

func AddItemsToken(items map[string]items.Item, headers http.Header) {
	expirationTime := time.Now().Add(JWTDuration)
	trainerItemsToken := &ItemsToken{
		Items:          items,
		ItemsHash:      generateHash(items),
		StandardClaims: jwt.StandardClaims{ExpiresAt: expirationTime.Unix()},
	}
	setTokenInHeader(ItemsTokenHeaderName, trainerItemsToken, headers)
}

func addTokenToHeader(headerName string, token interface{ jwt.Claims }, headers http.Header) {
	jwtToken := jwt.NewWithClaims(jwt.SigningMethodHS256, token)
	tokenString, err := jwtToken.SignedString(authJWTKey)
	if err != nil {
		panic(err)
	}

	headers.Add(headerName, tokenString)
}

func setTokenInHeader(headerName string, token interface{ jwt.Claims }, headers http.Header) {
	jwtToken := jwt.NewWithClaims(jwt.SigningMethodHS256, token)
	tokenString, err := jwtToken.SignedString(authJWTKey)
	if err != nil {
		panic(err)
	}
	headers.Set(headerName, tokenString)
}

func generateHash(toHash interface{}) string {
	marshaled, _ := json.Marshal(toHash)
	hash := md5.Sum(marshaled)
	encoder := base64.Encoding{}
	hashB64 := encoder.EncodeToString(hash[:])
	return hashB64
}
