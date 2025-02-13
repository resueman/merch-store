package converter

import (
	"github.com/resueman/merch-store/internal/entity"
	"github.com/resueman/merch-store/internal/model"
)

func ConvertPurchasesToInventory(purchases []entity.Purchase) []model.Inventory {
	inventory := make([]model.Inventory, 0, len(purchases))
	for _, purchase := range purchases {
		inventory = append(inventory, model.Inventory{
			Name:     purchase.Name,
			Quantity: purchase.Quantity,
		})
	}

	return inventory
}

// Мы должны аккаунт или username возвращать?
func ConvertTransfersToOutgoingTransfers(transfers []entity.Transfer) []model.OutgoingTransfer {
	outgoingTransfers := make([]model.OutgoingTransfer, 0, len(transfers))
	for _, transfer := range transfers {
		outgoingTransfers = append(outgoingTransfers, model.OutgoingTransfer{
			Amount:             transfer.Amount,
			RecipientAccountID: transfer.RecipientAccountID,
		})
	}

	return outgoingTransfers
}

// Мы должны аккаунт или username возвращать?
func ConvertTransfersToIncomingTransfers(transfers []entity.Transfer) []model.IncomingTransfer {
	incomingTransfers := make([]model.IncomingTransfer, 0, len(transfers))
	for _, transfer := range transfers {
		incomingTransfers = append(incomingTransfers, model.IncomingTransfer{
			Amount:          transfer.Amount,
			SenderAccountID: transfer.SenderAccountID,
		})
	}

	return incomingTransfers
}
