package main

import (
	"log"
	"net"
	"sync"

	"github.com/bestplay/wallet-service/internal/config"
	"github.com/bestplay/wallet-service/internal/handler"
	walletpb "github.com/bestplay/wallet-service/internal/proto"
	"github.com/bestplay/wallet-service/internal/service"
	"github.com/bestplay/wallet-service/internal/store"
	"github.com/gin-gonic/gin"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

func main() {
	cfg, err := config.Load("config/config.yaml")
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	var storeInstance store.Store
	switch cfg.Store.Type {
	case "mysql":
		log.Println("Using MySQL store")
		storeInstance, err = store.NewGormMySQLStore(cfg.Store.MySQL.DSN())
		if err != nil {
			log.Fatalf("Failed to create MySQL store: %v", err)
		}
	case "memory":
		fallthrough
	default:
		log.Println("Using memory store")
		storeInstance = store.NewMemoryStore()
	}

	svc := service.NewService(storeInstance)

	restHandler := handler.NewHandler(svc)
	grpcHandler := handler.NewGRPCHandler(svc)

	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		defer wg.Done()
		router := gin.Default()
		restHandler.RegisterRoutes(router)
		log.Println("REST server starting on :8080")
		if err := router.Run(":8080"); err != nil {
			log.Fatalf("REST server failed: %v", err)
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		lis, err := net.Listen("tcp", ":50051")
		if err != nil {
			log.Fatalf("failed to listen: %v", err)
		}
		s := grpc.NewServer()
		walletpb.RegisterWalletServiceServer(s, grpcHandler)
		// Enable reflection
		reflection.Register(s)
		log.Println("gRPC server starting on :50051")
		if err := s.Serve(lis); err != nil {
			log.Fatalf("gRPC server failed: %v", err)
		}
	}()

	wg.Wait()
}
