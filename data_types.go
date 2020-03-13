package NOVAPokemon

type User struct {
	Username string
	UID      string
}

type Trainer struct {
	UserUID  string
	BagUID   string
	Pokemons []Pokemon
	Level    int
	Coins    int
}

type Bag struct {
	UID   string
	Items []Item
}

type Item struct {
	UID   string
	Price int
}

type Pokemon struct {
	UID      string
	OwnerUID string
	Species  string
	Level    int
	HP       int
	Damage   int
}
