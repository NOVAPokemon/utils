package transactions

import (
	"github.com/NOVAPokemon/utils"
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
)

var transactionOfferMockup = utils.TransactionTemplate{
	Name:  "offer_1",
	Coins: 10,
	Price: 50,
}
var transactionRecordMockup = utils.TransactionRecord{
	TemplateName: transactionOfferMockup.Name,
	User:         "user1",
}

func TestMain(m *testing.M) {
	_ = RemoveAllTransactions()
	res := m.Run()
	_ = RemoveAllTransactions()

	os.Exit(res)
}

func TestAddTransaction(t *testing.T) {
	res, err := AddTransaction(transactionRecordMockup)

	if err != nil {
		t.Log(err)
		t.FailNow()
	}

	t.Log("added: " + res.Hex())
}

func TestGetTransactionsFromUser(t *testing.T) {
	added, err := AddTransaction(transactionRecordMockup)

	if err != nil {
		t.Log(err)
		t.FailNow()
		return
	}
	t.Log("added: " + added.Hex())

	transactions, err := GetTransactionsFromUser(transactionRecordMockup.User)

	if err != nil {
		t.Log(err)
		t.FailNow()
		return
	}

	contains := false

	for _, transaction := range transactions {
		if transaction.Id.Hex() == added.Hex() {
			contains = true
			break
		}
	}

	assert.True(t, contains)
}
