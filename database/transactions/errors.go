package transactions

import (
	"fmt"
	"github.com/pkg/errors"
)

const (
	errorRemoveAllTransactions = "error removeing all transactions"

	errorGetUserTransactionsFormat = "error getting user %s transactions"
	errorAddTransactionFormat = "error getting user %s transactions"
)

func wrapGetUserTransactionsError(err error, username string) error {
	return errors.Wrap(err, fmt.Sprintf(errorGetUserTransactionsFormat, username))
}

func wrapAddTransactionError(err error, username string) error {
	return errors.Wrap(err, fmt.Sprintf(errorAddTransactionFormat, username))
}

func wrapRemoveAllTransactionsError(err error) error {
	return errors.Wrap(err, errorRemoveAllTransactions)
}