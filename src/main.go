package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/gobwas/ws"
	"log"
	"net"
	"net/http"
	"syscall"
)

type ConnectionSet = map[net.Conn]struct{}
type Uint32Set = map[uint32]struct{}

var epoller, createEpollError = newEpoll()

var ConnectionRooms = map[net.Conn]Uint32Set{}
var RoomConnections = map[uint32]ConnectionSet{}

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

	ctx, _ := context.WithCancel(context.Background())
	go Loop(ctx)
	//go func() {
	//	time.Sleep(5 * time.Second)
	//	cancel()
	//}()

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

	// for every rooms that the user is in, delete it from the rooms
	if rooms, ok := ConnectionRooms[conn]; ok {
		for room, _ := range rooms {
			delete(RoomConnections[room], conn)

			if len(RoomConnections[room]) < 1 {
				delete(RoomConnections, room)
			}
		}
	}

	// delete it from connection to RoomConnections map
	delete(ConnectionRooms, conn)
	conn.Close()
}
