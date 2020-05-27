package api

import "fmt"

const GetTransactionOffersPath = "/microtransactions/offers/"
const GetPerformedTransactionsPath = "/microtransactions/transactions/"
const MakeTransactionPath = "/microtransactions/transactions/%s"

const OfferIdPathVar = "transactionId"

var GetTransactionOffersRoute = GetTransactionOffersPath
var GetPerformedTransactionsRoute = GetPerformedTransactionsPath
var MakeTransactionRoute = fmt.Sprintf(MakeTransactionPath, fmt.Sprintf("{%s}", OfferIdPathVar))
