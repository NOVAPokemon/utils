package api

import "fmt"

// PATHS

// trainer
const AddTrainerPath = "/trainers"
const GetTrainersPath = "/trainers"

const GetTrainerByUsernamePath = "/trainers/%s"
const UpdateTrainerStatsPath = "/trainers/%s"

// trainer pokemons
const AddPokemonPath = "/trainers/%s/pokemons/"
const RemovePokemonPath = "/trainers/%s/pokemons/%s"
const UpdatePokemonPath = "/trainers/%s/pokemons/%s"

// trainer bag
const AddItemToBagPath = "/trainers/%s/bag/"
const RemoveItemFromBagPath = "/trainers/%s/bag/%s"

// Tokens
const VerifyTrainerStatsPath = "/trainers/%s/stats/verify"
const VerifyPokemonsPath = "/trainers/%s/pokemons/verify"
const VerifyItemsPath = "/trainers/%s/bag/verify"

const GenerateAllTokensPath = "/trainers/%s/tokens"
const GenerateTrainerStatsTokenPath = "/trainers/%s/stats/token"
const GenerateItemsTokenPath = "/trainers/%s/items/token"
const GeneratePokemonsTokenPath = "/trainers/%s/pokemons/token"

const UpdateRegionPath = "/location"

// ROUTES
const UsernameVar = "username"
const PokemonIdVar = "pokemonId"
const ItemIdVar = "itemId"

const UsernameRouteVar = "{username}"
const PokemonIdRouteVar = "{pokemonId}"
const ItemIdRouteVar = "{itemId}"

var GetTrainerByUsernameRoute = fmt.Sprintf(GetTrainerByUsernamePath, UsernameRouteVar)
var UpdateTrainerStatsRoute = fmt.Sprintf(UpdateTrainerStatsPath, UsernameRouteVar)

// trainer pokemons
var AddPokemonRoute = fmt.Sprintf(AddPokemonPath, UsernameRouteVar)
var UpdatePokemonRoute = fmt.Sprintf(UpdatePokemonPath, UsernameRouteVar, PokemonIdRouteVar)
var RemovePokemonRoute = fmt.Sprintf(RemovePokemonPath, UsernameRouteVar, PokemonIdRouteVar)

// trainer bag
var AddItemToBagRoute = fmt.Sprintf(AddItemToBagPath, UsernameRouteVar)
var RemoveItemFromBagRoute = fmt.Sprintf(RemoveItemFromBagPath, UsernameRouteVar, ItemIdRouteVar)

// Tokens
var VerifyTrainerStatsRoute = fmt.Sprintf(VerifyTrainerStatsPath, UsernameRouteVar)
var VerifyPokemonRoute = fmt.Sprintf(VerifyPokemonsPath, UsernameRouteVar)
var VerifyItemsRoute = fmt.Sprintf(VerifyItemsPath, UsernameRouteVar)

var GenerateAllTokensRoute = fmt.Sprintf(GenerateAllTokensPath, UsernameRouteVar)
var GenerateTrainerStatsTokenRoute = fmt.Sprintf(GenerateTrainerStatsTokenPath, UsernameRouteVar)
var GenerateItemsTokenRoute = fmt.Sprintf(GenerateItemsTokenPath, UsernameRouteVar)
var GeneratePokemonsTokenRoute = fmt.Sprintf(GeneratePokemonsTokenPath, UsernameRouteVar)

var UpdateRegionRoute = UpdateRegionPath