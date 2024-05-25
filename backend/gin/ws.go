package main

import (
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"log"
	"net/http"
)

func WsUpgraderMiddleware() gin.HandlerFunc {
	up := websocket.Upgrader{}
	return func(ctx *gin.Context) {
		w, r := ctx.Writer, ctx.Request
		conn, err := up.Upgrade(w, r, nil)
		if err != nil {
			log.Println("connection upgrade error:", err)
			ctx.AbortWithStatus(http.StatusInternalServerError)
			return
		}
		ctx.Set("clientConn", conn)
		ctx.Next()
	}
}

func (s *Service) wsClientConnectionMiddleware() gin.HandlerFunc {
	return func(ctx *gin.Context) {

		clientChan := make(ClientChan)           // Initialize client channel
		s.stream.NewClients <- ClientConnection{ // Send new connection to event server
			Chn:  clientChan,
			Type: WsClient,
		}
		defer func() {
			s.stream.ClosedClients <- ClientConnection{ // Send closed connection to event server
				Chn:  clientChan,
				Type: WsClient,
			}
		}()
		ctx.Set("clientChan", clientChan)
		ctx.Next()
	}
}

func (s *Service) wsStreamHandler(ctx *gin.Context) {

	clientConn, ok := getContextConnection(ctx)
	if !ok || clientConn == nil {
		return
	}

	clientChan, ok := getContextClientChan(ctx)
	if !ok || clientChan == nil {
		return
	}

	//notify := ctx.Writer.CloseNotify()
	//go func() {
	//	<-notify
	//	s.logger.Info("ws client disconnected")
	//	clientConn.Close()
	//}()

	for {
		select {
		case msg := <-clientChan:
			if err := clientConn.WriteMessage(1, []byte(msg)); err != nil {
				s.logger.Debug("write message error: %s", err)
				clientConn.Close()
				return
			}
		}
	}

}
func getContextConnection(ctx *gin.Context) (*websocket.Conn, bool) {
	v, ok := ctx.Get("clientConn")
	if !ok {
		return nil, false
	}
	clientConn, ok := v.(*websocket.Conn)
	if !ok {
		return nil, false
	}
	return clientConn, true
}
func getContextClientChan(ctx *gin.Context) (ClientChan, bool) {
	v, ok := ctx.Get("clientChan")
	if !ok {
		return nil, false
	}
	clientChan, ok := v.(ClientChan)
	if !ok {
		return nil, false
	}
	return clientChan, true
}
