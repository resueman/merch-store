package integration

import (
	"net/http"
	"testing"

	"github.com/resueman/merch-store/internal/delivery/handlers/http/v1/converter"
	"github.com/resueman/merch-store/internal/model"
)

func TestBuyItem(t *testing.T) {
	defer cleanup()

	setup()

	// Auth (user, password) -> token
	token := authUser(t, "user", "password", http.StatusOK)

	accountInfo := &model.AccountInfo{
		Balance:           190,
		Inventory:         []model.Inventory{},
		IncomingTransfers: []model.IncomingTransfer{},
		OutgoingTransfers: []model.OutgoingTransfer{},
	}

	// GetInfo: after auth (token) -> expect initial balance, empty inventory, transactions
	expected := converter.ConvertAccountInfoToInfoResponse(accountInfo)

	getUserInfo(t, token, http.StatusOK, &expected)

	// BuyItem: too expensive item -> error
	buyItem(t, token, "powerbank", http.StatusBadRequest)

	// GetInfo: after try to buy too expensive item -> expect initial balance, no inventory, transfers
	getUserInfo(t, token, http.StatusOK, &expected)

	// BuyItem: first successful purchase, buy 1 book
	buyItem(t, token, "book", http.StatusOK) // cost(book) = 50

	// GetInfo: after purchasing book -> expect balance - cost(book), inventory = {book}, no transfers
	accountInfo.Balance -= 50
	accountInfo.Inventory = []model.Inventory{{Name: "book", Quantity: 1}}
	expected = converter.ConvertAccountInfoToInfoResponse(accountInfo)

	getUserInfo(t, token, http.StatusOK, &expected)

	// BuyItem: buy one more book -> success
	buyItem(t, token, "book", http.StatusOK)

	// GetInfo: after second purchase -> expect balance - 2 * cost(book), inventory = {book, book}, no transfers
	accountInfo.Balance -= 50
	accountInfo.Inventory = []model.Inventory{{Name: "book", Quantity: 2}}
	expected = converter.ConvertAccountInfoToInfoResponse(accountInfo)

	getUserInfo(t, token, http.StatusOK, &expected)

	// BuyItem: buy other item --  t-shirt
	buyItem(t, token, "t-shirt", http.StatusOK) // cost(t-shirt) = 80

	// GetInfo: after buying t-shirt ->
	// balance - 2 * cost(book) - cost(t-shirt),
	// inventory = {book, book, t-shirt}, no transfers
	accountInfo.Balance -= 80
	accountInfo.Inventory = []model.Inventory{{Name: "book", Quantity: 2}, {Name: "t-shirt", Quantity: 1}}
	expected = converter.ConvertAccountInfoToInfoResponse(accountInfo)

	getUserInfo(t, token, http.StatusOK, &expected)

	// BuyItem: purchase one more t-shirt, but it became too expensive -> error
	buyItem(t, token, "t-shirt", http.StatusBadRequest)

	// GetInfo: after try to buy too expensive t-shirt -> expect same balance and inventory, no transfers
	getUserInfo(t, token, http.StatusOK, &expected)

	// BuyItem: not existing item "jujuju" -> error
	buyItem(t, token, "jujuju", http.StatusBadRequest)

	// GetInfo: after not existing item -> expect same balance and inventory, no transfers
	getUserInfo(t, token, http.StatusOK, &expected)
}
