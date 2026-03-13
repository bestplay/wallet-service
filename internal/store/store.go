package store

import (
	"errors"

	"github.com/bestplay/wallet-service/internal/model"
	"github.com/shopspring/decimal"
)

var (
	ErrWalletNotFound      = errors.New("wallet not found")
	ErrInsufficientBalance = errors.New("insufficient balance")
)

type Store interface {
	CreateWallet(wallet *model.Wallet) error
	GetWallet(id string) (*model.Wallet, error)
	UpdateWallet(wallet *model.Wallet) error
	Transfer(sourceID, destID string, amount decimal.Decimal) error
	Recharge(walletID string, amount decimal.Decimal) error
}
