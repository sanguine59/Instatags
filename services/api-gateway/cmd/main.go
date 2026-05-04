package main

import (
	"log"
	"os"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	pbAI "github.com/sanguine59/img-getter-chrome-ext/protos/ai"
	"github.com/sanguine59/img-getter-chrome-ext/services/api-gateway/internal/handlers"
	"github.com/sony/gobreaker"
)

func main() {

	aiSvc := os.Getenv("AI_SERVICE_ADDR")
	if aiSvc == "" {
		aiSvc = "ai-service:50051"
	}

	aiConn, err := grpc.NewClient(aiSvc, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("Failed to connect to ai-service, %v", err)
	}
	defer aiConn.Close()

	aiClient := pbAI.NewAIServiceClient(aiConn)

	router := gin.Default()

	st := gobreaker.Settings{
		Name:        "AIServiceCB",
		Timeout:     60 * time.Second,
		Interval:    20 * time.Second,
		MaxRequests: 5,
		ReadyToTrip: func(count gobreaker.Counts) bool {
			return count.ConsecutiveFailures >= 3
		},
	}
	cb := gobreaker.NewCircuitBreaker(st)

	h := handlers.GatewayHandler(aiClient, cb)

	router.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:5173", "https://hbb.local", "http://hbb.local"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization", "Upgrade", "Connection"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	v1 := router.Group("api/v1")
	{
		upload := v1.Group("/upload")
		upload.POST("/send", h.PredictHashtags)
	}

	gatewayAddr := os.Getenv("GATEWAY_PORT")
	if gatewayAddr == "" {
		gatewayAddr = ":8000"
	}
	if err := router.Run(gatewayAddr); err != nil {
		log.Fatalf("Failed to start API Gateway: %v", err)
	}
}
