package model

import (
	"github.com/shopspring/decimal"
)

type Wallet struct {
	ID      string          `json:"id" gorm:"primaryKey;type:varchar(64)"`
	Balance decimal.Decimal `json:"balance" gorm:"type:decimal(20,4);not null;default:0.0000"`
}

func (Wallet) TableName() string {
	return "wallets"
}

type TransferRequest struct {
	SourceID string          `json:"sourceId"`
	DestID   string          `json:"destId"`
	Amount   decimal.Decimal `json:"amount"`
}

type RechargeRequest struct {
	WalletID string          `json:"walletId"`
	Amount   decimal.Decimal `json:"amount"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}
