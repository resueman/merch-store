package model

type AccountInfo struct {
	Balance           int
	Inventory         []Inventory
	IncomingTransfers []IncomingTransfer
	OutgoingTransfers []OutgoingTransfer
}
