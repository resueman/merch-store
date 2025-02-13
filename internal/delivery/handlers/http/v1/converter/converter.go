package converter

import (
	dto "github.com/resueman/merch-store/internal/api/v1"
	"github.com/resueman/merch-store/internal/model"
)

func ConvertAccountInfoToInfoResponse(info *model.AccountInfo) dto.InfoResponse {
	var infoResponse dto.InfoResponse
	infoResponse.Coins = &info.Balance

	if len(info.Inventory) > 0 {
		infoResponse.Inventory = convertInventory(info.Inventory)
	}

	if len(info.IncomingTransfers) > 0 || len(info.OutgoingTransfers) > 0 {
		infoResponse.CoinHistory = convertCoinHistory(info.IncomingTransfers, info.OutgoingTransfers)
	}

	return infoResponse
}

func convertInventory(inventory []model.Inventory) *[]struct {
	Quantity *int    `json:"quantity,omitempty"`
	Type     *string `json:"type,omitempty"`
} {
	result := make([]struct {
		Quantity *int    `json:"quantity,omitempty"`
		Type     *string `json:"type,omitempty"`
	}, len(inventory))

	for i, item := range inventory {
		result[i].Quantity = &item.Quantity
		result[i].Type = &item.Name
	}

	return &result
}

func convertCoinHistory(incoming []model.IncomingTransfer, outgoing []model.OutgoingTransfer) *struct {
	Received *[]struct {
		Amount   *int    `json:"amount,omitempty"`
		FromUser *string `json:"fromUser,omitempty"`
	} `json:"received,omitempty"`
	Sent *[]struct {
		Amount *int    `json:"amount,omitempty"`
		ToUser *string `json:"toUser,omitempty"`
	} `json:"sent,omitempty"`
} {
	history := &struct {
		Received *[]struct {
			Amount   *int    `json:"amount,omitempty"`
			FromUser *string `json:"fromUser,omitempty"`
		} `json:"received,omitempty"`
		Sent *[]struct {
			Amount *int    `json:"amount,omitempty"`
			ToUser *string `json:"toUser,omitempty"`
		} `json:"sent,omitempty"`
	}{}

	if len(incoming) > 0 {
		history.Received = convertIncomingTransfers(incoming)
	}

	if len(outgoing) > 0 {
		history.Sent = convertOutgoingTransfers(outgoing)
	}

	return history
}

func convertIncomingTransfers(transfers []model.IncomingTransfer) *[]struct {
	Amount   *int    `json:"amount,omitempty"`
	FromUser *string `json:"fromUser,omitempty"`
} {
	result := make([]struct {
		Amount   *int    `json:"amount,omitempty"`
		FromUser *string `json:"fromUser,omitempty"`
	}, len(transfers))

	for i, t := range transfers {
		result[i].Amount = &t.Amount
		result[i].FromUser = &t.SenderUsername
	}

	return &result
}

func convertOutgoingTransfers(transfers []model.OutgoingTransfer) *[]struct {
	Amount *int    `json:"amount,omitempty"`
	ToUser *string `json:"toUser,omitempty"`
} {
	result := make([]struct {
		Amount *int    `json:"amount,omitempty"`
		ToUser *string `json:"toUser,omitempty"`
	}, len(transfers))

	for i, t := range transfers {
		result[i].Amount = &t.Amount
		result[i].ToUser = &t.RecipientUsername
	}

	return &result
}
