package main

import (
	"fmt"
	"github.com/gorilla/websocket"
	"testing"
	"time"
)

func TestUnsubscribe(t *testing.T) {
	ws, done := SetupWebsocketServer(t)
	defer done()
	ConfirmEmptyServer(t)

	err := ws.WriteMessage(websocket.BinaryMessage, ConstructRoomPayload(Subscribe, 1, []byte{}))

	if err != nil {
		t.Errorf("Error sending subscribe message to websocket server: %s", err.Error())
	}
	time.Sleep(PROCESSING_DEADLINE_TIME)

	ws.WriteMessage(websocket.BinaryMessage, ConstructRoomPayload(Unsubscribe, 1, []byte{}))

	time.Sleep(PROCESSING_DEADLINE_TIME)

	if len(ConnectionRooms) != 1 {
		fmt.Println(ConnectionRooms)
		t.Errorf("Connection is not preserved in Conn -> Room[] map(ConnectionRooms) != 1")
	}

	if len(RoomConnections) != 0 {
		fmt.Println(RoomConnections)
		t.Errorf("After unsubscribing the room is not removed from RoomConnections map")
	}

}
