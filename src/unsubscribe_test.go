package main

import (
	"fmt"
	"github.com/gorilla/websocket"
	"testing"
	"time"
)

func TestUnsubscribe(t *testing.T) {
	ws, done := SetupWebsocketServer(t)

	err := ws.WriteMessage(websocket.BinaryMessage, ConstructRoomPayload(Subscribe, 1, []byte{}))

	if err != nil {
		t.Errorf("Error sending subscribe message to websocket server: %s", err.Error())
	}
	time.Sleep(PROCESSING_DEADLINE_TIME)

	ws.WriteMessage(websocket.BinaryMessage, ConstructRoomPayload(Unsubscribe, 1, []byte{}))

	time.Sleep(PROCESSING_DEADLINE_TIME)

	if len(ConnectionRooms) != 0 {
		fmt.Println(ConnectionRooms)
		t.Errorf("After unsubscribing user is not removed from ConnectionRooms map")
	}

	if len(RoomConnections) != 0 {
		fmt.Println(RoomConnections)
		t.Errorf("After unsubscribing the room is not removed from RoomConnections map")
	}

	done()
}
