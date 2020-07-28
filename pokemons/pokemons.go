package pokemons

type Pokemon struct {
	Id      string `json:"id" bson:"_id,omitempty"`
	Species string
	XP      float64
	Level   int
	HP      int
	MaxHP   int
	Damage  int
}
