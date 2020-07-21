package clients

type ClientDelays struct {
	Default     float64            `json:"default"`
	Multipliers map[string]float64 `json:"multipliers"`
}
