package items

import "github.com/pkg/errors"

var (
	ErrorNotAppliable = errors.New("item not appliable")
	ErrorInvalidId    = errors.New("invalid item id")
)
