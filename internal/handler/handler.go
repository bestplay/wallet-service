package handler

import (
	"errors"
	"net/http"

	"github.com/bestplay/wallet-service/internal/model"
	"github.com/bestplay/wallet-service/internal/service"
	"github.com/bestplay/wallet-service/internal/store"
	"github.com/gin-gonic/gin"
	"github.com/shopspring/decimal"
)

type Handler struct {
	service *service.Service
}

func NewHandler(service *service.Service) *Handler {
	return &Handler{
		service: service,
	}
}

func (h *Handler) RegisterRoutes(router *gin.Engine) {
	walletGroup := router.Group("/wallets")
	{
		walletGroup.POST("", h.CreateWallet)
		walletGroup.GET("/:id", h.GetWallet)
		walletGroup.POST("/transfer", h.Transfer)
		walletGroup.POST("/recharge", h.Recharge)
	}
}

func (h *Handler) CreateWallet(c *gin.Context) {
	wallet, err := h.service.CreateWallet()
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.ErrorResponse{Error: "failed to create wallet"})
		return
	}

	c.JSON(http.StatusCreated, wallet)
}

func (h *Handler) GetWallet(c *gin.Context) {
	id := c.Param("id")
	wallet, err := h.service.GetWallet(id)
	if err != nil {
		if errors.Is(err, store.ErrWalletNotFound) {
			c.JSON(http.StatusNotFound, model.ErrorResponse{Error: "wallet not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, model.ErrorResponse{Error: "failed to get wallet"})
		return
	}
	c.JSON(http.StatusOK, wallet)
}

func (h *Handler) Transfer(c *gin.Context) {
	var req model.TransferRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, model.ErrorResponse{Error: "invalid request body"})
		return
	}

	if req.Amount.LessThanOrEqual(decimal.Zero) {
		c.JSON(http.StatusBadRequest, model.ErrorResponse{Error: "amount must be positive"})
		return
	}

	err := h.service.Transfer(req.SourceID, req.DestID, req.Amount)
	if err != nil {
		switch {
		case errors.Is(err, store.ErrWalletNotFound):
			c.JSON(http.StatusNotFound, model.ErrorResponse{Error: "wallet not found"})
		case errors.Is(err, store.ErrInsufficientBalance):
			c.JSON(http.StatusBadRequest, model.ErrorResponse{Error: "insufficient balance"})
		default:
			c.JSON(http.StatusInternalServerError, model.ErrorResponse{Error: "transfer failed: "})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "success"})
}

func (h *Handler) Recharge(c *gin.Context) {
	var req model.RechargeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, model.ErrorResponse{Error: "invalid request body"})
		return
	}

	if req.Amount.LessThanOrEqual(decimal.Zero) {
		c.JSON(http.StatusBadRequest, model.ErrorResponse{Error: "amount must be positive"})
		return
	}

	err := h.service.Recharge(req.WalletID, req.Amount)
	if err != nil {
		if errors.Is(err, store.ErrWalletNotFound) {
			c.JSON(http.StatusNotFound, model.ErrorResponse{Error: "wallet not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, model.ErrorResponse{Error: "recharge failed"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "success"})
}
