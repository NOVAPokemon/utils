package errors

import "github.com/pkg/errors"

const (
	errorGetOffers             = "error getting offers"
	errorGetTransactionRecords = "error getting transaction records"
	errorPerformTransaction    = "error performing transaction"
)

func WrapGetOffersError(err error) error {
	return errors.Wrap(err, errorGetOffers)
}

func WrapGetTransactionsRecordsError(err error) error {
	return errors.Wrap(err, errorGetTransactionRecords)
}

func WrapPerformTransactionError(err error) error {
	return errors.Wrap(err, errorPerformTransaction)
}