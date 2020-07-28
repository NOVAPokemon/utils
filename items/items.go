package items

const (
	HealName       = "potion"
	ReviveName     = "revive"
	PokeBallName   = "poke-ball"
	MasterBallName = "master-ball"
)

var (
	HealItem = Item{
		Name:   HealName,
		Effect: HealEffect,
	}

	ReviveItem = Item{
		Name:   ReviveName,
		Effect: ReviveEffect,
	}

	PokeBallItem = Item{
		Name:   PokeBallName,
		Effect: PokeBallEffect,
	}

	MasterBallItem = Item{
		Name:   MasterBallName,
		Effect: MasterBallEffect,
	}
)

type Item struct {
	Id     string `json:"id" bson:"_id,omitempty"`
	Name   string
	Effect Effect
}

type StoreItem struct {
	Name  string
	Price int
}

func (storeItem StoreItem) ToItem() Item {
	return Item{
		Name:   storeItem.Name,
		Effect: GetEffectForItem(storeItem.Name),
	}
}

func (item Item) IsPokeBall() bool {
	switch item.Name {
	case PokeBallName:
		fallthrough
	case MasterBallName:
		return true
	default:
		return false
	}
}
