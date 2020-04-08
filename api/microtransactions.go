package api

import "fmt"

const GetTransactionOffersPath = "/offers/"
const GetPerformedTransactionsPath = "/transactions/"
const MakeTransactionPath = "/transactions/%s"

const OfferIdPathVar = "transactionId"

var GetTransactionOffersRoute = GetTransactionOffersPath
var GetPerformedTransactionsRoute = GetPerformedTransactionsPath
var MakeTransactionRoute = fmt.Sprintf(MakeTransactionPath, fmt.Sprintf("{%s}", OfferIdPathVar))
