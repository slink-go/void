package main

import (
	"github.com/gin-gonic/gin"
	"io"
)

func SseHeadersMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Content-Type", "text/event-stream")
		c.Writer.Header().Set("Cache-Control", "no-cache")
		c.Writer.Header().Set("Connection", "keep-alive")
		c.Writer.Header().Set("Transfer-Encoding", "chunked")
		c.Next()
	}
}

func (s *Service) sseClientConnectionMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		clientChan := make(ClientChan)           // Initialize client channel
		s.stream.NewClients <- ClientConnection{ // Send new connection to event server
			Chn:  clientChan,
			Type: SseClient,
		}
		defer func() {
			s.stream.ClosedClients <- ClientConnection{ // Send closed connection to event server
				Chn:  clientChan,
				Type: SseClient,
			}
		}()
		c.Set("clientChan", clientChan)
		c.Next()
	}
}

func (s *Service) sseStreamHandler(ctx *gin.Context) {

	clientChan, ok := getContextClientChan(ctx)
	if !ok || clientChan == nil {
		return
	}

	ctx.Stream(func(w io.Writer) bool {
		if msg, ok := <-clientChan; ok {
			ctx.SSEvent("message", msg) // Stream message to client from message channel
			return true
		}
		return false
	})
}
