package tokens

import (
	"crypto/md5"
	"encoding/base64"
	"encoding/json"
	originalHttp "net/http"
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
	b64Encoder    = base64.Encoding{}
)

func ExtractAndVerifyAuthToken(headers originalHttp.Header) (*AuthToken, error) {
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

func ExtractAndVerifyTrainerStatsToken(headers originalHttp.Header) (*TrainerStatsToken, error) {
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

func ExtractAndVerifyPokemonTokens(headers originalHttp.Header) ([]*PokemonToken, error) {
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

func ExtractAndVerifyItemsToken(headers originalHttp.Header) (*ItemsToken, error) {
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

func AddAuthToken(username string, headers originalHttp.Header) {
	expirationTime := time.Now().Add(JWTDuration)
	authToken := &AuthToken{
		Username:       username,
		StandardClaims: jwt.StandardClaims{ExpiresAt: expirationTime.Unix()},
	}

	setTokenInHeader(AuthTokenHeaderName, authToken, headers)
}

func AddTrainerStatsToken(stats utils.TrainerStats, headers originalHttp.Header) {
	expirationTime := time.Now().Add(JWTDuration)
	trainerStatsToken := &TrainerStatsToken{
		TrainerStats:   stats,
		TrainerHash:    GenerateHash(stats),
		StandardClaims: jwt.StandardClaims{ExpiresAt: expirationTime.Unix()},
	}
	setTokenInHeader(StatsTokenHeaderName, trainerStatsToken, headers)
}

func AddPokemonsTokens(pokemons map[string]pokemons.Pokemon, headers originalHttp.Header) {
	expirationTime := time.Now().Add(JWTDuration)
	for _, v := range pokemons {
		pokemonToken := &PokemonToken{
			Pokemon:        v,
			PokemonHash:    GenerateHash(v),
			StandardClaims: jwt.StandardClaims{ExpiresAt: expirationTime.Unix()},
		}
		addTokenToHeader(PokemonsTokenHeaderName, pokemonToken, headers)
	}
}

func AddItemsToken(items map[string]items.Item, headers originalHttp.Header) {
	expirationTime := time.Now().Add(JWTDuration)
	trainerItemsToken := &ItemsToken{
		Items:          items,
		ItemsHash:      GenerateHash(items),
		StandardClaims: jwt.StandardClaims{ExpiresAt: expirationTime.Unix()},
	}
	setTokenInHeader(ItemsTokenHeaderName, trainerItemsToken, headers)
}

func addTokenToHeader(headerName string, token interface{ jwt.Claims }, headers originalHttp.Header) {
	jwtToken := jwt.NewWithClaims(jwt.SigningMethodHS256, token)
	tokenString, err := jwtToken.SignedString(authJWTKey)
	if err != nil {
		panic(err)
	}

	headers.Add(headerName, tokenString)
}

func setTokenInHeader(headerName string, token interface{ jwt.Claims }, headers originalHttp.Header) {
	jwtToken := jwt.NewWithClaims(jwt.SigningMethodHS256, token)
	tokenString, err := jwtToken.SignedString(authJWTKey)
	if err != nil {
		panic(err)
	}
	headers.Set(headerName, tokenString)
}

func GenerateHash(toHash interface{}) string {
	marshaled, _ := json.Marshal(toHash)
	hash := md5.Sum(marshaled)
	hashB64 := b64Encoder.EncodeToString(hash[:])
	return hashB64
}
