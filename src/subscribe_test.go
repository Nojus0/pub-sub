package main

import (
	"encoding/binary"
	"github.com/gorilla/websocket"
	"testing"
	"time"
)

func ConstructRoomPayload(action Action, room uint32, data []byte) []byte {
	roomBytes := make([]byte, 4)

	binary.BigEndian.PutUint32(roomBytes, room)

	return append(append([]byte{action}, roomBytes...), data...)
}

func TestSubscribe(t *testing.T) {

	ws, done, _ := SetupWebsocketServer(t)
	defer done()
	ConfirmEmptyServer(t)

	err := ws.WriteMessage(websocket.BinaryMessage, ConstructRoomPayload(Subscribe, 0, []byte{}))

	if err != nil {
		t.Errorf("Error sending subscribe message to websocket server: %s", err.Error())
	}

	time.Sleep(PROCESSING_DEADLINE_TIME)

	subscriberAmount, roomAmount := len(ConnectionRooms), len(RoomConnections)

	if subscriberAmount != 1 {
		t.Errorf("User isn't added after %d ms deadline to the ConnectionRooms map", PROCESSING_DEADLINE_MS)
	}

	if roomAmount != 1 {
		t.Errorf("Room isn't added after %d ms deadline to the RoomConnections map", PROCESSING_DEADLINE_MS)
	}

}
