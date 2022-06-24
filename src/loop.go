package main

import (
	"context"
	"encoding/binary"
	"fmt"
	"github.com/gobwas/ws"
	"github.com/gobwas/ws/wsutil"
	"log"
	"net"
)

var EXISTS = struct{}{}

func Loop(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			fmt.Println("Epoll/Process Loop gracefully shutdown via context.Done()")
			return
		default:
			runLoop()
		}
	}

}

func runLoop() {
	connections, err := Epoller.Wait()
	if err != nil {
		log.Println("Failed to epoll wait: ", err.Error())
		return
	}

	for _, conn := range connections {
		if conn == nil {
			break
		}

		data, op, err := wsutil.ReadClientData(conn)

		if err != nil {
			if err := Epoller.Remove(conn); err != nil {
				log.Println("Failed to remove:", err.Error())
			}
			removeUser(conn)
			continue
		}

		if op == ws.OpClose {
			if err := Epoller.Remove(conn); err != nil {
				log.Println("Failed to remove:", err.Error())
			}
			removeUser(conn)
			continue
		}

		if len(data) < 5 {
			continue
		}

		var action = data[0]
		var roomId = binary.BigEndian.Uint32(data[1:5])

		switch action {
		case Subscribe:
			SubscribeAction(conn, roomId)
		case Publish:
			PublishAction(roomId, data)
		case Unsubscribe:
			UnsubscribeAction(conn, roomId)
		}
	}
}

func SubscribeAction(conn net.Conn, roomId uint32) {
	// * User Max Rooms Cap *
	if len(ConnectionRooms[conn]) > 50 {
		return
	}

	// * Check if already in the specified Room *
	if _, exists := ConnectionRooms[conn][roomId]; exists {
		return
	}

	// * If room doesn't exist, create it *
	if _, ok := RoomConnections[roomId]; !ok {
		RoomConnections[roomId] = ConnectionSet{}
	}
	// * If User doesn't have Room list create it *
	if _, ok := ConnectionRooms[conn]; !ok {
		ConnectionRooms[conn] = Uint32Set{}
	}

	// * Add room to user rooms map *
	ConnectionRooms[conn][roomId] = EXISTS

	// * Add Conn to room Set
	RoomConnections[roomId][conn] = EXISTS
}

func PublishAction(roomId uint32, data []byte) {

	// * If room doesn't exist *
	if _, ok := RoomConnections[roomId]; !ok {
		return
	}

	var payload = data[1:]
	var conns = RoomConnections[roomId]

	for conn := range conns {
		if err := wsutil.WriteServerMessage(conn, ws.OpBinary, payload); err != nil {
			log.Println("Failed to write message", err.Error())
		}
	}
}

func UnsubscribeAction(conn net.Conn, roomId uint32) {
	c := RoomConnections[roomId]
	delete(c, conn)

	if len(c) < 1 {
		delete(RoomConnections, roomId)
	} else {
		RoomConnections[roomId] = c
	}

	delete(ConnectionRooms[conn], roomId)
}
