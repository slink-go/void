package main

import (
	"fmt"
	"github.com/slink-go/logging"
	"sync"
	"time"
)

type ClientType int

const (
	SseClient ClientType = iota
	WsClient
)

type ClientConnection struct {
	Chn  chan string
	Type ClientType
}
type Stream struct {
	Message       chan string                // Events are pushed to this channel by the main events-gathering routine
	NewClients    chan ClientConnection      // New client connections
	ClosedClients chan ClientConnection      // Closed client connections
	TotalClients  map[chan string]ClientType // Total client connections
	logger        logging.Logger
}

type ClientChan chan string

func NewStreamingServer() (event *Stream) {
	event = &Stream{
		Message:       make(chan string),
		NewClients:    make(chan ClientConnection),
		ClosedClients: make(chan ClientConnection),
		TotalClients:  make(map[chan string]ClientType),
		logger:        logging.GetLogger("stream"),
	}
	go event.listen()
	return
}

func (stream *Stream) listen() {
	stream.logger.Info("start listening client events")
	for {
		select {
		case client := <-stream.NewClients: // Add new available client
			stream.TotalClients[client.Chn] = client.Type
			stream.logger.Info("Client added. %d registered clients", len(stream.TotalClients))
		case client := <-stream.ClosedClients: // Remove closed client
			delete(stream.TotalClients, client.Chn)
			close(client.Chn)
			stream.logger.Info("Client removed. %d registered clients", len(stream.TotalClients))
		case eventMsg := <-stream.Message: // Broadcast message to client
			for clientMessageChan := range stream.TotalClients {
				clientMessageChan <- eventMsg
			}
		}
	}
}

func (s *Service) StartDataGenerator(eventSource *Stream) {
	var once sync.Once
	once.Do(func() {
		s.startDataGenerator(eventSource)
	})
}
func (s *Service) startDataGenerator(eventSource *Stream) {
	go func() {
		var i int
		for {
			i++
			sseClientsCount := count(eventSource.TotalClients, SseClient)
			wsClientsCount := count(eventSource.TotalClients, WsClient)
			msg := fmt.Sprintf("%d - the time is %v (clients: sse:%d, ws:%d)", i, time.Now(), sseClientsCount, wsClientsCount)
			s.logger.Debug("generated new message: %s", msg)
			eventSource.Message <- msg
			time.Sleep(2 * time.Second)
		}
	}()
}

func count(clients map[chan string]ClientType, cType ClientType) int {
	result := 0
	for _, v := range clients {
		if v == cType {
			result++
		}
	}
	return result
}
