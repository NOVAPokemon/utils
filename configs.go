package utils

type ClientConfig struct {
	TradeConfig    TradesClientConfig   `json:"trades"`
	LocationConfig LocationClientConfig `json:"location"`
	BattleConfig   BattleClientConfig   `json:"battles"`
	RaidConfig     RaidConfig           `json:"raids"`
}

type TradesClientConfig struct {
	MaxItemsToTrade int `json:"max_items"`
	MaxSleepTime    int `json:"max_sleep_time"` // in seconds
}

type LocationClientConfig struct {
	UpdateInterval int                `json:"update_interval"` // in seconds
	Timeout        int                `json:"timeout"`
	Parameters     LocationParameters `json:"params"`
}

type BattleClientConfig struct {
	PokemonsPerBattle int `json:"pokemons_per_battle"`
}

type RaidConfig struct {
	PokemonsPerRaid int `json:"pokemons_per_raid"`
}

// Location parameters describe how the user will move. The moving probability indicates
// how likely is the user to move between location updates
// MaxMovingSpeed should be in meters per second
type LocationParameters struct {
	StartingLocation     Location `json:"starting_location"`
	MaxMovingSpeed       float64  `json:"max_moving_speed"`
	MovingProbability    float64  `json:"moving_probability"`
	MaxDistanceFromStart int      `json:"max_distance_from_start"`
}