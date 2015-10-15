// Chat client to talk to server.go. In order to run just user go run client.go and
// provide a username. This will listen for room events and display them on the
// console that is running client.go

package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"sync"
)

const (
	PORT               = ":65535"
	TYPE               = "tcp"
	HOST               = "localhost"
	DISCONNECT_MESSAGE = "You have disconnected from the server.\n"
)

var wGroup sync.WaitGroup

/*Function that handles user input, and sends to the server.*/
func HandleUserIn(connect net.Conn) {

	/*Have a reader and writer watch for user inputs.*/
	readWatcher := bufio.NewReader(os.Stdin)
	writeWatcher := bufio.NewWriter(connect)

	/*Infinite for loop to send to server, read the string then write the
	* string, then flush the writer.*/
	for {
		msg, errNo := readWatcher.ReadString('\n')
		if errNo != nil {
			fmt.Println(errNo)
			os.Exit(1)
		}
		/*Use an anonymous variable so we can just capture an error, if there
		* is one.*/
		_, errNo = writeWatcher.WriteString(msg)
		if errNo != nil {
			fmt.Println(errNo)
			os.Exit(1)
		}

		/*Flush the writer so we do not run into errors.*/
		errNo = writeWatcher.Flush()
		if errNo != nil {
			fmt.Println(errNo)
			os.Exit(1)
		}
	}
}

/*Function that handles server input, and sends it out to the console
* of every user.*/
func HandleServerIn(connect net.Conn) {
	read := bufio.NewReader(connect)
	/*Infinite loop for catching user input. */
	for {
		/*Watch for a new line, aka enter when the user sends a command.*/
		msg, errNo := read.ReadString('\n')
		/*Make sure no error was passed.*/
		if errNo != nil {
			fmt.Printf(DISCONNECT_MESSAGE)
			wGroup.Done()
			return
		}
		fmt.Print(msg)
	}

}

/*Start up both the functions of handle server and user in, then connect
* it to the server connection, which was defined as a constant at the
* top of the file.*/
func main() {
	/*Add a waitgroup, to wait for a routine to finish, much like a lock
	* or a semaphore*/
	wGroup.Add(1)

	connect, errNo := net.Dial(TYPE, HOST+PORT)
	if errNo != nil {
		fmt.Println(errNo)
	}
	/*Start both the handle for server and write ins.*/
	go HandleServerIn(connect)
	go HandleUserIn(connect)

	wGroup.Wait()
}
