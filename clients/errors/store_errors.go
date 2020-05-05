package errors

import "github.com/pkg/errors"

const (
	errorGetItems = "error getting items"
	errorBuyItem  = "error buying item"
)

func WrapGetItemsError(err error) error {
	return errors.Wrap(err, errorGetItems)
}

func WrapBuyItemError(err error) error {
	return errors.Wrap(err, errorBuyItem)
}
