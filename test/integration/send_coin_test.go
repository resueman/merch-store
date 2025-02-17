package integration

import (
	"net/http"
	"testing"

	"github.com/resueman/merch-store/internal/delivery/handlers/http/v1/converter"
	"github.com/resueman/merch-store/internal/model"
)

func TestSendCoin(t *testing.T) {
	defer cleanup()

	setup()

	accountA := &model.AccountInfo{
		Balance:           190,
		Inventory:         []model.Inventory{},
		IncomingTransfers: []model.IncomingTransfer{},
		OutgoingTransfers: []model.OutgoingTransfer{},
	}

	accountB := &model.AccountInfo{
		Balance:           190,
		Inventory:         []model.Inventory{},
		IncomingTransfers: []model.IncomingTransfer{},
		OutgoingTransfers: []model.OutgoingTransfer{},
	}

	users := map[string]*struct {
		password string
		token    string
		account  *model.AccountInfo
	}{
		"A": {
			password: "password_A",
			account:  accountA,
		},
		"B": {
			password: "password_B",
			account:  accountB,
		},
	}

	// Auth (A, password_A) -> tokenA
	users["A"].token = authUser(t, "A", users["A"].password, http.StatusOK)

	// Auth (B, password_B) -> tokenB
	users["B"].token = authUser(t, "B", users["B"].password, http.StatusOK)

	// SendCoin: A -> B: 10 -> success
	sendCoin(t, users["A"].token, "B", 10, http.StatusOK)

	// GetInfo: tokenA -> expect balance - 10, one outgoing transfer to B: 10
	users["A"].account.Balance -= 10
	users["A"].account.OutgoingTransfers = []model.OutgoingTransfer{{RecipientUsername: "B", Amount: 10}}
	expected := converter.ConvertAccountInfoToInfoResponse(users["A"].account)

	getUserInfo(t, users["A"].token, http.StatusOK, &expected)

	// GetInfo: tokenB -> expect balance + 10, one incoming transfer from A: 10
	users["B"].account.Balance += 10
	users["B"].account.IncomingTransfers = []model.IncomingTransfer{{SenderUsername: "A", Amount: 10}}
	expected = converter.ConvertAccountInfoToInfoResponse(users["B"].account)

	getUserInfo(t, users["B"].token, http.StatusOK, &expected)

	// SendCoin: A -> B: 100 -> success
	sendCoin(t, users["A"].token, "B", 100, http.StatusOK)

	// GetInfo: tokenA -> expect balance - 100 - 10, one two outgoing transfer to B: 100 and 10
	users["A"].account.Balance -= 100
	users["A"].account.OutgoingTransfers = []model.OutgoingTransfer{
		{RecipientUsername: "B", Amount: 100},
		{RecipientUsername: "B", Amount: 10},
	}
	expected = converter.ConvertAccountInfoToInfoResponse(users["A"].account)

	getUserInfo(t, users["A"].token, http.StatusOK, &expected)

	// GetInfo (tokenB) -> expect balance + 100 + 10, two incoming transfers from A: 100 and 10
	users["B"].account.Balance += 100
	users["B"].account.IncomingTransfers = []model.IncomingTransfer{
		{SenderUsername: "A", Amount: 100},
		{SenderUsername: "A", Amount: 10},
	}
	expected = converter.ConvertAccountInfoToInfoResponse(users["B"].account)

	getUserInfo(t, users["B"].token, http.StatusOK, &expected)

	// SendCoin: B -> A, 20 -> success
	sendCoin(t, users["B"].token, "A", 20, http.StatusOK)

	// GetInfo (tokenA) ->
	// balance - 100 - 10 + 20,
	// one incoming transfer from B: 20,
	// two outgoing transfers to B: 100 and 10
	users["A"].account.Balance += 20
	users["A"].account.IncomingTransfers = []model.IncomingTransfer{{SenderUsername: "B", Amount: 20}}
	users["A"].account.OutgoingTransfers = []model.OutgoingTransfer{
		{RecipientUsername: "B", Amount: 100},
		{RecipientUsername: "B", Amount: 10},
	}
	expectedA := converter.ConvertAccountInfoToInfoResponse(users["A"].account)

	getUserInfo(t, users["A"].token, http.StatusOK, &expectedA)

	// GetInfo (tokenB) ->
	// balance + 100 + 10 - 20,
	// two incoming transfers from A: 100 and 10,
	// one outgoing transfer to A: 20
	users["B"].account.Balance -= 20
	users["B"].account.IncomingTransfers = []model.IncomingTransfer{
		{SenderUsername: "A", Amount: 100},
		{SenderUsername: "A", Amount: 10},
	}
	users["B"].account.OutgoingTransfers = []model.OutgoingTransfer{{RecipientUsername: "A", Amount: 20}}
	expectedB := converter.ConvertAccountInfoToInfoResponse(users["B"].account)

	getUserInfo(t, users["B"].token, http.StatusOK, &expectedB)

	// SendCoin: A -> B, too much amount -> error
	sendCoin(t, users["A"].token, "B", 1000, http.StatusBadRequest)

	// GetInfo: tokenA -> expect same balance, same transfers
	getUserInfo(t, users["A"].token, http.StatusOK, &expectedA)

	// GetInfo: tokenB -> expect same balance, same transfers
	getUserInfo(t, users["B"].token, http.StatusOK, &expectedB)

	// SendCoin: A -> B, all coins -> success
	sendCoin(t, users["A"].token, "B", 100, http.StatusOK)

	// GetInfo: tokenA -> expect balance is 0, same transfers
	users["A"].account.Balance -= 100
	users["A"].account.OutgoingTransfers = []model.OutgoingTransfer{
		{RecipientUsername: "B", Amount: 100},
		{RecipientUsername: "B", Amount: 100},
		{RecipientUsername: "B", Amount: 10},
	}
	users["A"].account.IncomingTransfers = []model.IncomingTransfer{
		{SenderUsername: "B", Amount: 20},
	}

	expectedA = converter.ConvertAccountInfoToInfoResponse(users["A"].account)
	getUserInfo(t, users["A"].token, http.StatusOK, &expectedA)

	// GetInfo: tokenB -> expect balance, same transfers
	users["B"].account.Balance += 100
	users["B"].account.IncomingTransfers = []model.IncomingTransfer{
		{SenderUsername: "A", Amount: 100},
		{SenderUsername: "A", Amount: 100},
		{SenderUsername: "A", Amount: 10},
	}
	users["B"].account.OutgoingTransfers = []model.OutgoingTransfer{
		{RecipientUsername: "A", Amount: 20},
	}
	expectedB = converter.ConvertAccountInfoToInfoResponse(users["B"].account)
	getUserInfo(t, users["B"].token, http.StatusOK, &expectedB)

	// SendCoin: A -> B, 1 -> error because A has 0 balance
	sendCoin(t, users["A"].token, "B", 1, http.StatusBadRequest)

	// GetInfo: tokenA -> expect balance is 0, same transfers
	getUserInfo(t, users["A"].token, http.StatusOK, &expectedA)

	// GetInfo: tokenB -> expect balance, same transfers
	getUserInfo(t, users["B"].token, http.StatusOK, &expectedB)

}
