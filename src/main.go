package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net"
	"net/http"
	"syscall"
	"time"

	"github.com/gobwas/ws"
)

type ConSet = map[net.Conn]struct{}

var epoller, createEpollError = newEpoll()
var ConnectionRooms = map[net.Conn][]uint32{}

var RoomConnections = map[uint32]ConSet{}

func wsHandler(w http.ResponseWriter, r *http.Request) {
	// Upgrade connection
	conn, _, _, err := ws.UpgradeHTTP(r, w)
	if err != nil {
		return
	}

	if err := epoller.Add(conn); err != nil {
		log.Println("Failed to add connection:", err.Error())
		removeUser(conn)
	}
}

func main() {
	var rLimit syscall.Rlimit
	if err := syscall.Getrlimit(syscall.RLIMIT_NOFILE, &rLimit); err != nil {
		panic(err)
	}
	rLimit.Cur = rLimit.Max
	fmt.Println(rLimit.Max)
	if err := syscall.Setrlimit(syscall.RLIMIT_NOFILE, &rLimit); err != nil {
		panic(err)
	}

	// Loop epoll
	if createEpollError != nil {
		panic(createEpollError)
	}

	ctx, cancel := context.WithCancel(context.Background())
	go Loop(ctx)
	go func() {
		time.Sleep(5 * time.Second)
		cancel()
	}()

	http.HandleFunc("/ws", wsHandler)

	http.HandleFunc("/", func(rw http.ResponseWriter, r *http.Request) {
		rw.Write([]byte("Pub sub server"))
	})

	port := flag.Uint("port", 8080, "port to listen on")
	flag.Parse()

	fmt.Println("Started listening on port", *port)
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

func removeUser(conn net.Conn) {
	connectedRooms, exists := ConnectionRooms[conn]

	// for every room that the user is in, delete it from the room
	if exists {
		for _, userRoom := range connectedRooms {
			delete(RoomConnections[userRoom], conn)

			if len(RoomConnections[userRoom]) < 1 {
				delete(RoomConnections, userRoom)
			}
		}
	}

	// delete it from connection to RoomConnections map
	delete(ConnectionRooms, conn)
	conn.Close()
}
