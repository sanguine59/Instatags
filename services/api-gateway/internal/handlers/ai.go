package handlers

import (
	"bytes"
	"context"
	"errors"
	"io"
	"log"
	"net/http"
	"time"

	pbAI "github.com/sanguine59/img-getter-chrome-ext/protos/ai"
	"github.com/gin-gonic/gin"
	"github.com/sony/gobreaker"
)

func (h *Gateway) PredictHashtags(c *gin.Context) {
	
	file, _, err := c.Request.FormFile("image")

	
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Image file is required"})
		return
	}
	defer file.Close()

	
	buf := bytes.NewBuffer(nil)
	if _, err := io.Copy(buf, file); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read image file"})
		return
	}

	
	
	
	result, err := h.CB.Execute(func() (interface{}, error) {
		ctx, cancel := context.WithTimeout(c.Request.Context(), 20*time.Second)
		defer cancel()

		return h.AIClient.PredictHashtags(ctx, &pbAI.PredictRequest{
			ImageData: buf.Bytes(),
		})
	})

	if err != nil {
		log.Printf("AI Service Error: %v", err)

		if err == gobreaker.ErrOpenState {
			c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Service unavaible, try again later"})
			return
		}

		if errors.Is(err, context.DeadlineExceeded) {
			c.JSON(http.StatusGatewayTimeout, gin.H{"error": "Request timed out"})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate hashtags"})
		return
	}

	resp, ok := result.(*pbAI.PredictResponse)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal type error"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"hashtags": resp.Hashtags,
	})

}
