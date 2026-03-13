package store

import (
	"sync"

	"github.com/bestplay/wallet-service/internal/model"
	"github.com/shopspring/decimal"
)

type MemoryStore struct {
	mu      sync.RWMutex
	wallets map[string]*model.Wallet
}

func NewMemoryStore() *MemoryStore {
	return &MemoryStore{
		wallets: make(map[string]*model.Wallet),
	}
}

func (s *MemoryStore) CreateWallet(wallet *model.Wallet) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.wallets[wallet.ID] = wallet
	return nil
}

func (s *MemoryStore) GetWallet(id string) (*model.Wallet, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	wallet, ok := s.wallets[id]
	if !ok {
		return nil, ErrWalletNotFound
	}
	return wallet, nil
}

func (s *MemoryStore) UpdateWallet(wallet *model.Wallet) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.wallets[wallet.ID]; !ok {
		return ErrWalletNotFound
	}
	s.wallets[wallet.ID] = wallet
	return nil
}

func (s *MemoryStore) Transfer(sourceID, destID string, amount decimal.Decimal) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	source, ok := s.wallets[sourceID]
	if !ok {
		return ErrWalletNotFound
	}

	dest, ok := s.wallets[destID]
	if !ok {
		return ErrWalletNotFound
	}

	if source.Balance.LessThan(amount) {
		return ErrInsufficientBalance
	}

	source.Balance = source.Balance.Sub(amount)
	dest.Balance = dest.Balance.Add(amount)

	s.wallets[sourceID] = source
	s.wallets[destID] = dest

	return nil
}

func (s *MemoryStore) Recharge(walletID string, amount decimal.Decimal) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	wallet, ok := s.wallets[walletID]
	if !ok {
		return ErrWalletNotFound
	}

	wallet.Balance = wallet.Balance.Add(amount)
	s.wallets[walletID] = wallet

	return nil
}
