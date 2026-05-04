package handlers

import (
	pbAI "github.com/sanguine59/img-getter-chrome-ext/protos/ai"
	"github.com/sony/gobreaker"
)

type Gateway struct {
	AIClient pbAI.AIServiceClient
	CB       *gobreaker.CircuitBreaker
}

func GatewayHandler(aiClient pbAI.AIServiceClient, cb *gobreaker.CircuitBreaker) *Gateway {
	return &Gateway{
		AIClient: aiClient,
		CB:       cb,
	}
}
