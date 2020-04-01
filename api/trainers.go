package api

import "fmt"

// PATHS

// trainer
const AddTrainerPath = "/trainers/"
const GetTrainersPath = "/trainers/"

const GetTrainerByUsernamePath = "/trainers/%s"
const UpdateTrainerStatsPath = "/trainers/%s"

// trainer pokemons
const AddPokemonPath = "/trainers/%s/pokemons/"
const RemovePokemonPath = "/trainers/%s/pokemons/%s"

// trainer bag
const AddItemToBagPath = "/trainers/%s/bag/"
const RemoveItemFromBagPath = "/trainers/%s/bag/%s"

// Tokens
const VerifyTrainerStatsPath = "/trainers/%s/stats/verify"
const VerifyPokemonPath = "/trainers/%s/pokemons/%s/verify"

const VerifyItemsPath = "/trainers/%s/bag/verify"
const GenerateAllTokensPath = "/trainers/%s/tokens"
const GenerateTrainerStatsTokenPath = "/trainers/%s/stats/token"
const GenerateItemsTokenPath = "/trainers/%s/items/token"
const GeneratePokemonsTokenPath = "/trainers/%s/pokemons/token"

// ROUTES
const UsernameVar = "username"
const PokemonIdVar = "pokemonId"
const ItemIdVar = "itemId"

var GetTrainerByUsernameRoute = fmt.Sprintf(GetTrainerByUsernamePath, UsernameVar)
var UpdateTrainerStatsRoute = fmt.Sprintf(UpdateTrainerStatsPath, UsernameVar)

// trainer pokemons
var AddPokemonRoute = fmt.Sprintf(AddPokemonPath, UsernameVar)
var RemovePokemonRoute = fmt.Sprintf(RemovePokemonPath, UsernameVar, PokemonIdVar)

// trainer bag
var AddItemToBagRoute = fmt.Sprintf(AddItemToBagPath, UsernameVar)
var RemoveItemFromBagRoute = fmt.Sprintf(RemoveItemFromBagPath, UsernameVar, ItemIdVar)

// Tokens
var VerifyTrainerStatsRoute = fmt.Sprintf(VerifyTrainerStatsPath, UsernameVar)
var VerifyPokemonRoute = fmt.Sprintf(VerifyPokemonPath, UsernameVar, PokemonIdVar)
var VerifyItemsRoute = fmt.Sprintf(VerifyItemsPath, UsernameVar)

var GenerateAllTokensRoute = fmt.Sprintf(GenerateAllTokensPath, UsernameVar)
var GenerateTrainerStatsTokenRoute = fmt.Sprintf(GenerateTrainerStatsTokenPath, UsernameVar)
var GenerateItemsTokenRoute = fmt.Sprintf(GenerateItemsTokenPath, UsernameVar)
var GeneratePokemonsTokenRoute = fmt.Sprintf(GeneratePokemonsTokenPath, UsernameVar)
