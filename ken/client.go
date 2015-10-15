package main

import (
	"sync"
	"net"	
	"fmt"
	"os"
	"bufio"
)

const (
	CONN_PORT = ":1234"
	CONN_TYPE = "tcp"

	MSG_DISCONNECT = "Disconnected.\n"
)

var wg sync.WaitGroup

// Reads from the socket and outputs to the console.
func Read(conn net.Conn) {
	reader := bufio.NewReader(conn)
	for {
		str, err := reader.ReadString('\n')
		if err != nil {
			fmt.Printf(MSG_DISCONNECT)
			wg.Done()
			return
		}
		fmt.Print(str)
	}
}

// read stdin to socket
func Write(conn net.Conn) {
	reader := bufio.NewReader(os.Stdin)
	writer := bufio.NewWriter(conn)

	for {
		str, err := reader.ReadString('\n')
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		_, err = writer.WriteString(str)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		err = writer.Flush()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
	}
}

// creates read and write thread, onnects to server via socket
func main() {
	wg.Add(1)

	conn, err := net.Dial(CONN_TYPE, CONN_PORT)
	if err != nil {
		fmt.Println(err)
	}

	go Read(conn)
	go Write(conn)

	wg.Wait()
}