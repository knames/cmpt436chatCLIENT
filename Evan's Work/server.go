// Simple chat server in go language. Will update as I continue to work with TCP and telnet
// connections.
package main

// Require specific packages.  Util will mainly consist of all client structures, 
// an error checker, encoding and decoding characters, logging, etc.
// A major amount of work will be done within utils.
import (
	"bufio"
	"fmt"
	"io"
	"log"
	"net"
	"strings"
	"time"
	"regexp"
	"./util"
)

const MAINLOBBY = "lobby"

func main() 
{
	// Possible we can make a file to load properites.
	props := uti.LoadConfig()
	// This is for the tcp sockets.
	pSocket, pError := net.Listen("tcp", ":", props.PortNum)
	util.CheckForError(pError, "Cannot create a server!")

	// Have some output stating whether server gets started.
	fmt.Printf("Chat server has begun on port %v...\n", props.PortNum)

	// Server has to have a simple loop to keep running, it will forever
	// listen until a user joins, then will wait for user input.
	for
	{
		// First, we accept new connections.
		newConn, errNum := pSocket.Accept()
		
		// Check for any connection issues.
		util.CheckForError(errNum, "Cannot accpet new connections at this time.")

		// Let's make sure to keep track of client details to see rooms or anything
		// of the similar things. Let's send them into the main lobby first.
		client := util.Client{Connection: conn, Room: MAINLOBBY, Properties: props}
		// Now register the client.
		client.Register();

		// Make the client request non-blocking so we don't run into issues.
		mainChannel := make(chan string)
		go waitForUserIn(mainChannel, &client)
		go handleUserInput(mainChannel, &client, props)

		// Tell the client we are ready to accept anything they want to do.
		util.SendClientMessage("Chat Ready!", properties.PortNum, &client, true, props)
	}
}

// This is the function to wait for user input which is buffered by \n
// and signal the main channel.
func waitForUserIn(output chan string, client *util.Client)
{
	defer close(output)

	clientReader := bufio.NewReader(client.Connection)
	for
	{
		curLine, errNo := reader.ReadBytes('\n')
		if errNo != nil
		{
			// If there is no username, remove the client
			// from the list.
			client.Close(true);
			return
		}
	output <- string(curLine)
	}
}

// Now we can listen for user input, and handle in specific cases. For our current assignment 1
// we can use creation of rooms, joining, list, and sending messages to a room, and leave rooms.
func HandleUserInput(input <-chan string, client *util.Client, props util.Properties)
{
	for
	{
		curMessage := <- input
		// Check if message is not blank.
		if (curMessage != "")
		{
		curMessage = strings.TrimSpace(curMessage)
		// Parse out messages.
		curAction, body := getAction(message)

		// After white space trimming, and getting the action (join, leave, etc)
		// let's start using case statements.
		if (curMessage != "")
		{
			switch curAction
			{
				// user sends a message.
				case "message":
					util.SendClientMessage("message", body, client, false, props)
				// user provides their username.
				case "user":
					util.SendClientMessage("connect", "", client, false, props)

				// The user disconnects.
				case "disconnect":
					client.Close(false);

				// User enters a room.
				case "enter":
					// Make sure body (anything after /case is not empty.
					if (body != "")
					{
						client.Room = body
						util.SendClientMessage("enter", body, client, false, props)
					}

				// User leaves the current room.
				case "leave":
					// Check if room is not the main lobby.
					if (client.Room != MAINLOBBY)
					{
						util.SendClientMessage("leave", client.Room, client, false, props)
						client.Room = MAINLOBBY
					}
				default:
					util.SendClientMessage("unrecognized", action, client, true, props)
			}
		}
	}
}

// Now we parse out the message contents to return individual values.
func getAction(message string) (string, string)
{
	thisRegEx, _ := regexp.Compile(`^\/([^\s]*_\s*(.*)$`)
	result := thisRegEx.FindAllStringSubmatch(message, -1)
	// If length is one, then we return the results in individual values.
	if (len(result) == 1)
	{
		return result[0][1], result[0][2]
	}
	// Otherwise return a blank statement.
	return "", ""c
}