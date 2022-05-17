package main

import (
	"encoding/binary"
	"flag"
	"fmt"
	"log"
	"net"
	"net/http"
	"syscall"

	"github.com/gobwas/ws"
	"github.com/gobwas/ws/wsutil"
)

type ConSet = map[net.Conn]struct{}

var epoller *epoll
var subscriptions = map[net.Conn][]uint32{}

var rooms = map[uint32]ConSet{}

func wsHandler(w http.ResponseWriter, r *http.Request) {
	// Upgrade connection
	conn, _, _, err := ws.UpgradeHTTP(r, w)
	if err != nil {
		return
	}

	if err := epoller.Add(conn); err != nil {
		log.Printf("Failed to add connection %v", err)
		removeUser(conn)
	}
}

func main() {
	var rLimit syscall.Rlimit
	if err := syscall.Getrlimit(syscall.RLIMIT_NOFILE, &rLimit); err != nil {
		panic(err)
	}
	rLimit.Cur = rLimit.Max
	if err := syscall.Setrlimit(syscall.RLIMIT_NOFILE, &rLimit); err != nil {
		panic(err)
	}

	// Start epoll
	var err error
	epoller, err = newEpoll()
	if err != nil {
		panic(err)
	}

	go Start()
	http.HandleFunc("/ws", wsHandler)

	http.HandleFunc("/", func(rw http.ResponseWriter, r *http.Request) {
		rw.Write([]byte("Pub sub server"))
	})

	port := flag.Uint("port", 8080, "port to listen on")
	flag.Parse()

	fmt.Printf("Started listening on port %d", *port)
	if err := http.ListenAndServe(fmt.Sprintf("0.0.0.0:%d", *port), nil); err != nil {
		log.Fatal(err)
	}
}

type Action = uint8

const (
	Subscribe   Action = 0
	Publish     Action = 1
	Unsubscribe Action = 2
)

func Start() {
	for {
		connections, err := epoller.Wait()
		if err != nil {
			log.Printf("Failed to epoll wait %v", err)
			continue
		}
		for _, conn := range connections {
			if conn == nil {
				break
			}
			if data, op, err := wsutil.ReadClientData(conn); err != nil {
				if err := epoller.Remove(conn); err != nil {
					log.Printf("Failed to remove %v", err)
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
					if len(subscriptions[conn]) > 50 {
						break choose
					}

					// * If room doesn't exist, create it *
					if rooms[roomId] == nil {
						rooms[roomId] = ConSet{}
					}

					for _, s := range subscriptions[conn] {
						if s == roomId {
							break choose
						}
					}

					subscriptions[conn] = append(subscriptions[conn], roomId)

					// * Add Conn to room Set
					rooms[roomId][conn] = struct{}{}
				case Publish:
					if rooms[roomId] == nil {
						break choose
					}

					var payload = append([]byte{data[1], data[2], data[3], data[4]}, data[5:]...)
					var roomConns = rooms[roomId]

					for conn := range roomConns {
						if err := wsutil.WriteServerMessage(conn, ws.OpText, payload); err != nil {
							log.Printf("Failed to write message %v", err)
						}
					}

				case Unsubscribe:
					delete(rooms, roomId)
				}
			}
		}
	}
}

func removeUser(conn net.Conn) {
	connectedRooms, exists := subscriptions[conn]

	// for every room that the user is in, delete it from the room
	if exists {
		for _, userRoom := range connectedRooms {
			delete(rooms[userRoom], conn)

			if len(rooms[userRoom]) < 1 {
				delete(rooms, userRoom)
			}
		}
	}

	// delete it from connection to rooms map
	delete(subscriptions, conn)
	conn.Close()
}
