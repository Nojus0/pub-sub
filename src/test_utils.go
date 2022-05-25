package main

import (
	"context"
	"github.com/gorilla/websocket"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

type CleanupFunc = func()

var PROCESSING_DEADLINE_MS time.Duration = 16
var PROCESSING_DEADLINE_TIME = PROCESSING_DEADLINE_MS * time.Millisecond

func SetupWebsocketServer(t *testing.T) (*websocket.Conn, CleanupFunc) {

	ctx, cancel := context.WithCancel(context.Background())
	go Loop(ctx)

	s := httptest.NewServer(http.HandlerFunc(wsHandler))

	uri := "ws" + strings.TrimPrefix(s.URL, "http")

	ws, _, err := websocket.DefaultDialer.Dial(uri, nil)

	if err != nil {
		t.Errorf("Error occurred while connecting to test websocket server: %s", err.Error())
	}

	return ws, func() {
		cancel()
		s.Close()
		ws.Close()
	}
}
