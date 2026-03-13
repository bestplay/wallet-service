package service

import (
	"errors"

	"github.com/bestplay/wallet-service/internal/model"
	"github.com/bestplay/wallet-service/internal/store"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type Service struct {
	store store.Store
}

func NewService(store store.Store) *Service {
	return &Service{
		store: store,
	}
}

func (s *Service) CreateWallet() (*model.Wallet, error) {
	id := uuid.Must(uuid.NewV7()).String()

	wallet := &model.Wallet{
		ID:      id,
		Balance: decimal.Zero,
	}

	if err := s.store.CreateWallet(wallet); err != nil {
		return nil, err
	}

	return wallet, nil
}

func (s *Service) GetWallet(id string) (*model.Wallet, error) {
	return s.store.GetWallet(id)
}

func (s *Service) Transfer(sourceID, destID string, amount decimal.Decimal) error {
	if sourceID == destID {
		return errors.New("source and dest must be different")
	}
	if amount.LessThanOrEqual(decimal.Zero) {
		return errors.New("amount must be positive")
	}

	return s.store.Transfer(sourceID, destID, amount)
}

func (s *Service) Recharge(walletID string, amount decimal.Decimal) error {
	if amount.LessThanOrEqual(decimal.Zero) {
		return errors.New("amount must be positive")
	}

	return s.store.Recharge(walletID, amount)
}
