package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"sync"
)

var (
	clients   = make(map[net.Conn]string)
	mu        sync.Mutex
	broadcast = make(chan string)
)

func main() {
	port := ":4444"
	listener, err := net.Listen("tcp", port)

	if err != nil {
		fmt.Println("ERROR: Server failed to start! - ", err)
		os.Exit(1)
	}
	defer listener.Close()

	fmt.Println("Starting server at port " + port)

	go handleBroadcast()

	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("ERROR: Connection not accepted! - ", err)
			continue
		}

		go handleClient(conn)
	}
}

func handleClient(conn net.Conn) {
	defer conn.Close()

	conn.Write([]byte("Username: "))
	name, err := bufio.NewReader(conn).ReadString('\n')
	if err != nil {
		fmt.Println("ERROR: Couldn't catch username - ", err)
		return
	}
	name = name[:len(name)-1]

	mu.Lock()
	clients[conn] = name
	mu.Unlock()

	fmt.Printf("Connected client: %s (%s)\n", name, conn.RemoteAddr())
	broadcast <- fmt.Sprintf("%s joined the chat!", name)

	scanner := bufio.NewScanner(conn)
	for scanner.Scan() {
		msg := scanner.Text()
		broadcast <- fmt.Sprintf("[%s]: %s", name, msg)
	}

	mu.Lock()
	delete(clients, conn)
	mu.Unlock()
	fmt.Printf("Client disconnected: %s\n", name)
	broadcast <- fmt.Sprintf("%s left the chat.", name)
}

func handleBroadcast() {
	for {
		msg := <-broadcast
		fmt.Println(msg)

		mu.Lock()
		for conn := range clients {
			_, err := conn.Write([]byte(msg + "\n"))
			if err != nil {
				fmt.Println("ERROR: Could not send message! - ", err)
			}
		}
		mu.Unlock()
	}
}
