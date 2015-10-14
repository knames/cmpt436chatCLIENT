// Simple chat server in go language. Will update as I continue to work with TCP and telnet
// connections.
package main

// Require specific packages.  Util will mainly consist of all client structures,
// an error checker, encoding and decoding characters, logging, etc.
// A major amount of work will be done within utils.
import (
	"./util"
	"bufio"
	"fmt"
	"net"
	"regexp"
	"strings"
)

// Array of rooms to list.
var rooms = [200]string{"lobby"}

const MAINLOBBY = "lobby"

func main() {
	// Possible we can make a file to load properites.
	props := util.LoadConfig()
	// This is for the tcp sockets.
	pSocket, pError := net.Listen("tcp", ":"+props.Port)
	util.CheckForError(pError, "Cannot create a server!")

	// Have some output stating whether server gets started.
	fmt.Printf("Chat server %v has begun on port %v...\n", props.Host, props.Port)

	// Server has to have a simple loop to keep running, it will forever
	// listen until a user joins, then will wait for user input.
	for {
		// First, we accept new connections.
		newConn, errNum := pSocket.Accept()

		// Check for any connection issues.
		util.CheckForError(errNum, "Cannot accpet new connections at this time.")

		// Let's make sure to keep track of client details to see rooms or anything
		// of the similar things. Let's send them into the main lobby first.
		client := util.Client{UserConnection: newConn, Room: MAINLOBBY, Prop: props}
		// Now register the client.
		client.Register()

		// Make the client request non-blocking so we don't run into issues.
		mainChannel := make(chan string)
		go waitForUserIn(mainChannel, &client)
		go HandleUserInput(mainChannel, &client, props)

		// Tell the client we are ready to accept anything they want to do.
		util.SendClientMessage("ready", props.Port, &client, true, props)
	}
}

// This is the function to wait for user input which is buffered by \n
// and signal the main channel.
func waitForUserIn(output chan string, client *util.Client) {
	defer close(output)
	// fmt.Printf("Output is now closed.")

	clientReader := bufio.NewReader(client.UserConnection)
	for {
		curLine, errNo := clientReader.ReadBytes('\n')
		if errNo != nil {
			// If there is no username, remove the client
			// from the list.
			client.Close(true)
			return
		}
		output <- string(curLine)
	}
}

// Now we can listen for user input, and handle in specific cases. For our current assignment 1
// we can use creation of rooms, joining, list, and sending messages to a room, and leave rooms.
func HandleUserInput(input <-chan string, client *util.Client, props util.Properties) {
	for {
		curMessage := <-input
		// Check if message is not blank.
		if curMessage != "" {
			curMessage = strings.TrimSpace(curMessage)
			// Parse out messages.
			curAction, body := getAction(curMessage)

			// After white space trimming, and getting the action (join, leave, etc)
			// let's start using case statements.
			if curAction != "" {
				switch curAction {
				// user sends a message.
				case "message":
					util.SendClientMessage("message", body, client, false, props)
				// user provides their username.
				case "user":
					client.User = body
					util.SendClientMessage("connect", "", client, false, props)

				// The user disconnects.
				case "disconnect":
					client.Close(false)

				// User enters a room.
				case "enter":
					// Make sure body (anything after /case is not empty.
					if body != "" {
						client.Room = body
						//fmt.Printf("%s", rooms[0])
						util.SendClientMessage("enter", body, client, false, props)
						for i := 0; i < 200; i++ {
							if rooms[i] != "" {
								rooms[i] = client.Room
								fmt.Printf("%s\n", rooms[i])
							}
						}
					}
				// User wants to list all current rooms.
				case "list":
					for i := 0; i < 200; i++ {
						if rooms[i] != "" {
							util.SendClientMessage("list", rooms[i], client, false, props)
						}
						//fmt.Printf("Did not find room.\n")
					}
				// Print out the list of rooms.

				// User leaves the current room.
				case "leave":
					// Check if room is not the main lobby.
					if client.Room != MAINLOBBY {
						util.SendClientMessage("leave", client.Room, client, false, props)
						client.Room = MAINLOBBY
					}
				default:
					util.SendClientMessage("unrecognized", curAction, client, true, props)
				}
			}
		}
	}
}

// Now we parse out the message contents to return individual values.
func getAction(message string) (string, string) {
	actionRegex, _ := regexp.Compile(`^\/([^\s]*)\s*(.*)$`)
	res := actionRegex.FindAllStringSubmatch(message, -1)
	if len(res) == 1 {
		return res[0][1], res[0][2]
	}
	return "", ""
}
