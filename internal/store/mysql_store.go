package store

import (
	"fmt"

	"github.com/bestplay/wallet-service/internal/model"
	"github.com/shopspring/decimal"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type GormMySQLStore struct {
	db *gorm.DB
}

func NewGormMySQLStore(dsn string) (*GormMySQLStore, error) {
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
		SkipDefaultTransaction: true,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to MySQL: %w", err)
	}

	createTableSQL := `
	CREATE TABLE IF NOT EXISTS wallets (
		id VARCHAR(64) PRIMARY KEY,
		balance DECIMAL(20, 4) NOT NULL DEFAULT 0.0000
	) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
	`
	if err := db.Exec(createTableSQL).Error; err != nil {
		return nil, fmt.Errorf("failed to create table: %w", err)
	}

	return &GormMySQLStore{db: db}, nil
}

func (s *GormMySQLStore) CreateWallet(wallet *model.Wallet) error {
	if err := s.db.Create(wallet).Error; err != nil {
		return fmt.Errorf("failed to create wallet: %w", err)
	}
	return nil
}

func (s *GormMySQLStore) GetWallet(id string) (*model.Wallet, error) {
	var wallet model.Wallet
	if err := s.db.First(&wallet, "id = ?", id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, ErrWalletNotFound
		}
		return nil, fmt.Errorf("failed to get wallet: %w", err)
	}
	return &wallet, nil
}

func (s *GormMySQLStore) UpdateWallet(wallet *model.Wallet) error {
	result := s.db.Model(&model.Wallet{}).Where("id = ?", wallet.ID).Update("balance", wallet.Balance)
	if result.Error != nil {
		return fmt.Errorf("failed to update wallet: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return ErrWalletNotFound
	}
	return nil
}

func (s *GormMySQLStore) Transfer(sourceID, destID string, amount decimal.Decimal) error {
	return s.db.Transaction(func(tx *gorm.DB) error {
		var sourceWallet model.Wallet
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&sourceWallet, "id = ?", sourceID).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				return ErrWalletNotFound
			}
			return fmt.Errorf("failed to get source wallet: %w", err)
		}

		if sourceWallet.Balance.LessThan(amount) {
			return ErrInsufficientBalance
		}

		var destWallet model.Wallet
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&destWallet, "id = ?", destID).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				return ErrWalletNotFound
			}
			return fmt.Errorf("failed to get dest wallet: %w", err)
		}

		sourceWallet.Balance = sourceWallet.Balance.Sub(amount)
		destWallet.Balance = destWallet.Balance.Add(amount)

		if err := tx.Save(&sourceWallet).Error; err != nil {
			return fmt.Errorf("failed to update source wallet: %w", err)
		}

		if err := tx.Save(&destWallet).Error; err != nil {
			return fmt.Errorf("failed to update dest wallet: %w", err)
		}

		return nil
	})
}

func (s *GormMySQLStore) Recharge(walletID string, amount decimal.Decimal) error {
	return s.db.Transaction(func(tx *gorm.DB) error {
		var wallet model.Wallet
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&wallet, "id = ?", walletID).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				return ErrWalletNotFound
			}
			return fmt.Errorf("failed to get wallet: %w", err)
		}

		wallet.Balance = wallet.Balance.Add(amount)

		if err := tx.Save(&wallet).Error; err != nil {
			return fmt.Errorf("failed to update wallet: %w", err)
		}

		return nil
	})
}

func (s *GormMySQLStore) Close() error {
	sqlDB, err := s.db.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}
