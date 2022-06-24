package main

import (
	"fmt"
	"github.com/gorilla/websocket"
	"testing"
	"time"
)

func TestDisconnect(t *testing.T) {
	ws, done, _ := SetupWebsocketServer(t)
	defer done()
	ConfirmEmptyServer(t)

	ws.WriteMessage(websocket.BinaryMessage, ConstructRoomPayload(Subscribe, 1, []byte{}))

	time.Sleep(PROCESSING_DEADLINE_TIME)
	ws.Close()
	time.Sleep(PROCESSING_DEADLINE_TIME)

	if len(ConnectionRooms) != 0 {
		fmt.Println(ConnectionRooms)
		t.Errorf("Connection is not preserved in Conn -> Room[] map(ConnectionRooms) != 1")
	}

	if len(RoomConnections) != 0 {
		fmt.Println(RoomConnections)
		t.Errorf("After unsubscribing the room is not removed from RoomConnections map")
	}

}
