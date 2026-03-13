package handler

import (
	"context"
	"errors"

	walletpb "github.com/bestplay/wallet-service/internal/proto"
	"github.com/bestplay/wallet-service/internal/service"
	"github.com/bestplay/wallet-service/internal/store"
	"github.com/shopspring/decimal"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type GRPCHandler struct {
	walletpb.UnimplementedWalletServiceServer
	service *service.Service
}

func NewGRPCHandler(service *service.Service) *GRPCHandler {
	return &GRPCHandler{
		service: service,
	}
}

func (h *GRPCHandler) CreateWallet(ctx context.Context, req *walletpb.CreateWalletRequest) (*walletpb.Wallet, error) {
	wallet, err := h.service.CreateWallet()
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to create wallet")
	}
	return &walletpb.Wallet{
		Id:      wallet.ID,
		Balance: wallet.Balance.String(),
	}, nil
}

func (h *GRPCHandler) GetWallet(ctx context.Context, req *walletpb.GetWalletRequest) (*walletpb.Wallet, error) {
	wallet, err := h.service.GetWallet(req.Id)
	if err != nil {
		if errors.Is(err, store.ErrWalletNotFound) {
			return nil, status.Errorf(codes.NotFound, "wallet not found")
		}
		return nil, status.Errorf(codes.Internal, "failed to get wallet")
	}
	return &walletpb.Wallet{
		Id:      wallet.ID,
		Balance: wallet.Balance.String(),
	}, nil
}

func (h *GRPCHandler) Transfer(ctx context.Context, req *walletpb.TransferRequest) (*walletpb.TransferResponse, error) {
	amount, err := decimal.NewFromString(req.Amount)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid amount")
	}
	if amount.LessThanOrEqual(decimal.Zero) {
		return nil, status.Errorf(codes.InvalidArgument, "amount must be positive")
	}

	err = h.service.Transfer(req.SourceId, req.DestId, amount)
	if err != nil {
		switch {
		case errors.Is(err, store.ErrWalletNotFound):
			return nil, status.Errorf(codes.NotFound, "wallet not found")
		case errors.Is(err, store.ErrInsufficientBalance):
			return nil, status.Errorf(codes.InvalidArgument, "insufficient balance")
		case err.Error() == "source and dest must be different":
			return nil, status.Errorf(codes.InvalidArgument, "source and dest must be different")
		default:
			return nil, status.Errorf(codes.Internal, "transfer failed")
		}
	}

	return &walletpb.TransferResponse{Status: "success"}, nil
}

func (h *GRPCHandler) Recharge(ctx context.Context, req *walletpb.RechargeRequest) (*walletpb.RechargeResponse, error) {
	amount, err := decimal.NewFromString(req.Amount)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid amount")
	}
	if amount.LessThanOrEqual(decimal.Zero) {
		return nil, status.Errorf(codes.InvalidArgument, "amount must be positive")
	}

	err = h.service.Recharge(req.WalletId, amount)
	if err != nil {
		if errors.Is(err, store.ErrWalletNotFound) {
			return nil, status.Errorf(codes.NotFound, "wallet not found")
		}
		return nil, status.Errorf(codes.Internal, "recharge failed")
	}

	return &walletpb.RechargeResponse{Status: "success"}, nil
}
