package main

import (
	"context"
	"encoding/binary"
	"fmt"
	"github.com/gobwas/ws"
	"github.com/gobwas/ws/wsutil"
	"log"
)

func Loop(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			fmt.Println("Loop gracefully shutdown via context.Done()")
			return
		default:
			runLoop()
		}
	}

}

func runLoop() {
	connections, err := epoller.Wait()
	if err != nil {
		log.Println("Failed to epoll wait: ", err.Error())
		return
	}

	for _, conn := range connections {
		if conn == nil {
			break
		}
		if data, op, err := wsutil.ReadClientData(conn); err != nil {
			if err := epoller.Remove(conn); err != nil {
				log.Println("Failed to remove:", err.Error())
			}
			removeUser(conn)
		} else {

			if op == ws.OpClose {
				epoller.Remove(conn)
				removeUser(conn)
				continue
			}

			if len(data) < 5 {
				continue
			}

			var action Action = data[0]
			var roomId = binary.BigEndian.Uint32(data[1:5])

		choose:
			switch action {
			case Subscribe:

				// * User Max Rooms Cap *
				if len(ConnectionRooms[conn]) > 50 {
					break choose
				}

				// * If room doesn't exist, create it *
				if RoomConnections[roomId] == nil {
					RoomConnections[roomId] = ConSet{}
				}

				for _, s := range ConnectionRooms[conn] {
					if s == roomId {
						break choose
					}
				}

				ConnectionRooms[conn] = append(ConnectionRooms[conn], roomId)

				// * Add Conn to room Set
				RoomConnections[roomId][conn] = struct{}{}
			case Publish:
				if RoomConnections[roomId] == nil {
					break choose
				}

				var payload = append([]byte{data[1], data[2], data[3], data[4]}, data[5:]...)
				var roomConns = RoomConnections[roomId]

				for conn := range roomConns {
					if err := wsutil.WriteServerMessage(conn, ws.OpText, payload); err != nil {
						log.Println("Failed to write message", err.Error())
					}
				}

			case Unsubscribe:
				c := RoomConnections[roomId]
				delete(c, conn)

				if len(c) < 1 {
					delete(RoomConnections, roomId)
				} else {
					RoomConnections[roomId] = c
				}

				if rooms, ok := ConnectionRooms[conn]; ok {
					newRooms := make([]uint32, 0)
					for _, rr := range rooms {
						if rr != roomId {
							newRooms = append(newRooms, rr)
						}
					}
					ConnectionRooms[conn] = newRooms
				}

			}
		}
	}
}
