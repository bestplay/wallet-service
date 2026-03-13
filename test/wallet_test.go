package integration

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"

	"github.com/bestplay/wallet-service/internal/handler"
	"github.com/bestplay/wallet-service/internal/model"
	"github.com/bestplay/wallet-service/internal/service"
	"github.com/bestplay/wallet-service/internal/store"

	"github.com/gin-gonic/gin"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
)

func setupTestServer() (*service.Service, *gin.Engine) {
	gin.SetMode(gin.TestMode)
	store := store.NewMemoryStore()
	service := service.NewService(store)
	handler := handler.NewHandler(service)

	router := gin.Default()
	handler.RegisterRoutes(router)

	return service, router
}

func TestWalletConcurrency(t *testing.T) {
	_, router := setupTestServer()

	var wg sync.WaitGroup
	const numGoroutines = 100
	walletIDs := make([]string, numGoroutines)
	var mu sync.Mutex

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			w := httptest.NewRecorder()
			req, _ := http.NewRequest("POST", "/wallets", nil)
			router.ServeHTTP(w, req)

			assert.Equal(t, http.StatusCreated, w.Code)

			var resp model.Wallet
			json.Unmarshal(w.Body.Bytes(), &resp)

			mu.Lock()
			walletIDs[idx] = resp.ID
			mu.Unlock()
		}(i)
	}

	wg.Wait()

	for i := 0; i < numGoroutines; i++ {
		assert.NotEmpty(t, walletIDs[i])
	}
}

func TestRechargeConcurrency(t *testing.T) {
	_, router := setupTestServer()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/wallets", nil)
	router.ServeHTTP(w, req)

	var walletResp model.Wallet
	json.Unmarshal(w.Body.Bytes(), &walletResp)
	walletID := walletResp.ID

	var wg sync.WaitGroup
	const numRecharges = 100
	rechargeAmount := decimal.NewFromFloat(10.0)

	for i := 0; i < numRecharges; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()

			payload := map[string]interface{}{
				"walletId": walletID,
				"amount":   10.0,
			}
			body, _ := json.Marshal(payload)

			w := httptest.NewRecorder()
			req, _ := http.NewRequest("POST", "/wallets/recharge", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")
			router.ServeHTTP(w, req)

			assert.Equal(t, http.StatusOK, w.Code)
		}()
	}

	wg.Wait()

	w = httptest.NewRecorder()
	req, _ = http.NewRequest("GET", fmt.Sprintf("/wallets/%s", walletID), nil)
	router.ServeHTTP(w, req)

	var finalWallet model.Wallet
	json.Unmarshal(w.Body.Bytes(), &finalWallet)

	expected := rechargeAmount.Mul(decimal.NewFromInt(numRecharges))
	assert.True(t, finalWallet.Balance.Equal(expected),
		"Expected balance %s, got %s", expected, finalWallet.Balance)
}

func TestTransferConcurrency(t *testing.T) {
	service, router := setupTestServer()

	sourceWallet, _ := service.CreateWallet()
	destWallet, _ := service.CreateWallet()

	// 初始余额为 0，所以充值 10000 方便测试
	service.Recharge(sourceWallet.ID, decimal.NewFromFloat(10000.0))

	var wg sync.WaitGroup
	const numTransfers = 100
	transferAmount := decimal.NewFromFloat(10.0)

	for i := 0; i < numTransfers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()

			payload := map[string]interface{}{
				"sourceId": sourceWallet.ID,
				"destId":   destWallet.ID,
				"amount":   10.0,
			}
			body, _ := json.Marshal(payload)

			w := httptest.NewRecorder()
			req, _ := http.NewRequest("POST", "/wallets/transfer", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")
			router.ServeHTTP(w, req)
		}()
	}

	wg.Wait()

	finalSource, _ := service.GetWallet(sourceWallet.ID)
	finalDest, _ := service.GetWallet(destWallet.ID)

	// 源钱包：10000 - 1000 = 9000
	expectedSource := decimal.NewFromFloat(10000.0).Sub(transferAmount.Mul(decimal.NewFromInt(numTransfers)))
	// 目标钱包：0 + 1000 = 1000
	expectedDest := transferAmount.Mul(decimal.NewFromInt(numTransfers))

	assert.True(t, finalSource.Balance.Equal(expectedSource),
		"Source balance mismatch: expected %s, got %s", expectedSource, finalSource.Balance)
	assert.True(t, finalDest.Balance.Equal(expectedDest),
		"Dest balance mismatch: expected %s, got %s", expectedDest, finalDest.Balance)
}
