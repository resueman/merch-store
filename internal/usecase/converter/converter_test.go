package converter

import (
	"testing"

	"github.com/resueman/merch-store/internal/entity"
	"github.com/resueman/merch-store/internal/model"
	"github.com/stretchr/testify/require"
)

func TestConvertPurchasesToInventory(t *testing.T) {
	tests := []struct {
		name      string
		purchases []entity.Purchase
		expected  []model.Inventory
	}{
		{
			name:      "empty input",
			purchases: []entity.Purchase{},
			expected:  []model.Inventory{},
		},
		{
			name: "non-empty input",
			purchases: []entity.Purchase{
				{Name: "Item1", Quantity: 10},
				{Name: "Item2", Quantity: 5},
			},
			expected: []model.Inventory{
				{Name: "Item1", Quantity: 10},
				{Name: "Item2", Quantity: 5},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			inventory := ConvertPurchasesToInventory(tt.purchases)
			require.Equal(t, tt.expected, inventory)
		})
	}
}

func TestConvertTransfersToOutgoingTransfers(t *testing.T) {
	tests := []struct {
		name      string
		transfers []entity.Transfer
		expected  []model.OutgoingTransfer
	}{
		{
			name:      "empty input",
			transfers: []entity.Transfer{},
			expected:  []model.OutgoingTransfer{},
		},
		{
			name: "non-empty input",
			transfers: []entity.Transfer{
				{Amount: 100, RecipientUsername: "user1"},
				{Amount: 200, RecipientUsername: "user2"},
			},
			expected: []model.OutgoingTransfer{
				{Amount: 100, RecipientUsername: "user1"},
				{Amount: 200, RecipientUsername: "user2"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			outgoingTransfers := ConvertTransfersToOutgoingTransfers(tt.transfers)
			require.Equal(t, tt.expected, outgoingTransfers)
		})
	}
}

func TestConvertTransfersToIncomingTransfers(t *testing.T) {
	tests := []struct {
		name      string
		transfers []entity.Transfer
		expected  []model.IncomingTransfer
	}{
		{
			name:      "empty input",
			transfers: []entity.Transfer{},
			expected:  []model.IncomingTransfer{},
		},
		{
			name: "non-empty input",
			transfers: []entity.Transfer{
				{Amount: 150, SenderUsername: "user3"},
				{Amount: 250, SenderUsername: "user4"},
			},
			expected: []model.IncomingTransfer{
				{Amount: 150, SenderUsername: "user3"},
				{Amount: 250, SenderUsername: "user4"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			incomingTransfers := ConvertTransfersToIncomingTransfers(tt.transfers)
			require.Equal(t, tt.expected, incomingTransfers)
		})
	}
}
