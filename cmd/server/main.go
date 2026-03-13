package main

import (
	"github.com/bestplay/wallet-service/internal/handler"
	"github.com/bestplay/wallet-service/internal/service"
	"github.com/bestplay/wallet-service/internal/store"
	"github.com/gin-gonic/gin"
)

func main() {
	router := gin.Default()
	service := service.NewService(store.NewMemoryStore())
	handler := handler.NewHandler(service)

	handler.RegisterRoutes(router)

	router.Run(":8080")
}
