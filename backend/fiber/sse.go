package main

import (
	"bufio"
	"fmt"
	"github.com/gofiber/fiber/v2"
	"github.com/valyala/fasthttp"
	"time"
)

func (s *Service) sseTestHandler(c *fiber.Ctx) error {
	return c.Render("sse-test", fiber.Map{
		//"url": fmt.Sprintf("/api/%s/sse", s.applicationId),
		"url": "/api/sse",
	})
}

func (s *Service) sseHandler(c *fiber.Ctx) error {

	eventChan := make(chan string)
	s.sseChannels[eventChan] = struct{}{} // Add the client to the clients map

	c.Set("Content-Type", "text/event-stream")
	c.Set("Cache-Control", "no-cache")
	c.Set("Connection", "keep-alive")
	c.Set("Transfer-Encoding", "chunked")

	c.Status(fiber.StatusOK).Context().SetBodyStreamWriter(s.getSseStreamWriter(c, eventChan))

	return nil
}

func (s *Service) getSseStreamWriter(c *fiber.Ctx, chn chan string) fasthttp.StreamWriter {
	return fasthttp.StreamWriter(func(w *bufio.Writer) {
		defer func() {
			delete(s.sseChannels, chn) // Remove the client when they disconnect
			close(chn)
		}()
		fmt.Println("WRITER")
		for {
			select {
			case msg := <-chn:
				if msg != "" {
					s.logger.Debug("write message to client: '%s'", msg)
					fmt.Fprintf(w, "data: Message: %s\n\n", msg)
					fmt.Println(msg)
					err := w.Flush()
					if err != nil {
						// Refreshing page in web browser will establish a new
						// SSE connection, but only (the last) one is alive, so
						// dead connections must be closed here.
						fmt.Printf("Error while flushing: %v. Closing http connection.\n", err)
						return
					}
				}
			case <-c.Context().Done():
				return
			}
		}
	})
}

func (s *Service) StartDataGenerator() {
	go func() {
		var i int
		for {
			i++
			msg := fmt.Sprintf("%d - the time is %v (clients: #%d)", i, time.Now(), len(s.sseChannels))
			s.logger.Debug("generated new message: %s", msg)
			s.broadcast(msg)
			time.Sleep(2 * time.Second)
		}
	}()
}
func (s *Service) broadcast(data string) {
	for chn, _ := range s.sseChannels {
		s.logger.Info("broadcast message '%s' to %v", data, chn)
		chn <- data
	}
}
