package main

import (
	"github.com/gorilla/websocket"
	"testing"
	"time"
)

func TestSend(t *testing.T) {
	ws, done, uri := SetupWebsocketServer(t)
	defer done()
	ConfirmEmptyServer(t)

	ws1, _, _ := websocket.DefaultDialer.Dial(uri, nil)
	defer ws1.Close()

	ws1.WriteMessage(websocket.BinaryMessage, ConstructRoomPayload(Subscribe, 1, []byte{}))
	ws.WriteMessage(websocket.BinaryMessage, ConstructRoomPayload(Publish, 1, []byte("test message")))

	op, data, err := ws1.ReadMessage()
	if err != nil {
		t.Errorf("Error while reading message. %s\n", err.Error())
	}

	if op != websocket.BinaryMessage {
		t.Errorf("Recieved message is not binary(opcode) got: %d\n", op)
	}

	msg := string(data[4:])

	if msg != "test message" {
		t.Errorf("Recieved wrong message: %s\n", msg)
	}

	time.Sleep(PROCESSING_DEADLINE_TIME)

}
