// Chat client to talk to server.go. In order to run just user go run client.go and
// provide a username. This will listen for room events and display them on the
// console that is running client.go

package main

import (
	"fmt"
	"os"
	"net"
	"bufio"
	"regexp"
	"strings"
	"./util"
)

// Look for a command that involves a /
var stdInMsgRegex, _ = regexp.Compile(`^\/([^\s]*)\s*(.*)$`)

// Look for specific chat commands that involves users.
var chatServRespRegex, _ = regexp.Compile(`^\/([^\s]*)\s(?:\[([^\]]*)\]?\s*(.*)$`)

// Make a structure for Command details, may need the Command, username and body of the
// Command.
type Command struct {
	//Use for the mentioned commands in server such as message leave join etc.
	Cmd, User, Body string

}

// Main portion of program now. Will just watch for chat server and user input commands.
func main() {
	username, props := getConfig();

	// Connect to the server.
	connect, errNum := net.Dial("tcp", props.Host + ":" + props.Port)
	// Error check.
	util.CheckForError(errNum, "Connection Refused")
	defer connect.Close()

	// Listen for both commands.
	go watchForServerIn(username, props, connect)
	for true {
		watchForConsoleIn(connect)
	}
}

// We have to keep watch for console input. We need to listen for a message 
// that gets sent to the server.
func watchForConsoleIn(connect net.Conn) {
	// Create a new reader for standard input.
	readWatcher := bufio.NewReader(os.Stdin)
	
	// Now we wait for commands and listen. Use case statements to make this easy
	// to catch errors.
	for true {
		// Look for error checks, see if user has dropped connection.
		msg, errNo := readWatcher.ReadString('\n')
		util.CheckForError(errNo, "Lost Connection To Console")

		// Trim trailing and leading spaces, so we don't get a bad case.
		msg = strings.TrimSpace(msg)

		// Check if message is not empty (aka if the user just sent " ")
		if (msg != "") {
			// State the new input being parsed as command.
			command := parseInput(msg)
			// Uncomment for debug
			// fmt.Printf("%q", command)
			
			// Check if there is a blank message, if so just send it as a blank command.
			if (command.Cmd == "") {
				sendCommandToServ("message", msg, connect);
			} else {
				switch command.Cmd {
					// If user disconnects.
					case "disconnect":
						sendCommandToServ("disconnect", "", connect)
					// If user leaves a room.
					case "leave":
						sendCommandToServ("leave", "", connect)
					// If user enters room.
					case "enter":
						sendCommandToServ("enter", command.Body, connect)
					// If user wants to list rooms.
					//case "list":
					//	sendCommandToServ("list", "", connect)
					// Default case is unknown commands.
					default:
						fmt.Printf("Unknown command: \"%s\"\n", command.Cmd)
				}
			}
		}
	}
}

// Now we have to listen for commands that come from the chat server.
// Keep a constant loop running until we receive a command, then filter
// it by case to see whether its a valid command sent from the server,
// then we will print it out. Such as a user entering/leaving a room and
// disconnecting.
// This function is much like the previous one, but relies a lot more
// on the server rather than the client.

func watchForServerIn(user string, props util.Properties, connect net.Conn) {
	readIn := bufio.NewReader(connect)

	for true {
		// Check for error. If lost connection, exit.
		msg, errNo := readIn.ReadString('\n')
		util.CheckForError(errNo, "Lost connection to Server!");
		msg = strings.TrimSpace(msg)
		
		// If message is blank, can't do anything so check if non-blank.
		if (msg != "") {
			Cmd := parseCommand(msg)
			switch Cmd.Cmd {
				// Check if user is sending a message to another user.
				case "message":
					// If username doesn't equal itself.
					if (Cmd.User != user) {
						fmt.Printf(props.ReceivedMsg + "\n", Cmd.User, Cmd.Body)
					}
				// Initial we are ready, sends out username to server.
				case "ready":
					sendCommandToServ("user", user, connect)

				// Handle connect and disconnect calls.
				case "connect":
					fmt.Printf(props.HasEnteredLobbyMsg + "\n", Cmd.User)

				// Handle connect and disconnect calls.
				case "disconnect":
					fmt.Printf(props.HasLeftLobbyMsg + "\n", Cmd.User)

				// Handle cases of leaving and entering rooms now.
				case "enter":
					fmt.Printf(props.HasEnteredRoomMsg + "\n", Cmd.User, Cmd.Body)
					//have a list of rooms and add to that list.
					// Maybe through it in util file.

				// Handle cases of leaving and entering rooms now.
				case "leave":
					fmt.Printf(props.HasLeftRoomMsg + "\n", Cmd.User, Cmd.Body)
			}
		}
	}
}

// Send a command to the chat server, we just simply write the message to the net connection.
// Just send it as a byte array(?)
func sendCommandToServ(cmd string, body string, connect net.Conn) {
	// Encode the message so we can send to server.
	msg := fmt.Sprintf("/%v %v\n", util.Encode(cmd), util.Encode(body));
	connect.Write([]byte(msg))
}

// This command is used to parse input message and return a command structure.
// Note it will only have command and body.
func parseInput(msg string) Command  {
	response := stdInMsgRegex.FindAllStringSubmatch(msg, -1)
	// Check if the length is 1, that means there's a command.
	if (len(response) == 1) {
		return Command {
			Cmd: response[0][1],
			Body: response[0][2],
		}
	} else {
		return Command {
			Body: util.Decode(msg),
		}
	}
}

// Another parse command to look for specific things that involve usernames.
// Command at top means we're returning command.
func parseCommand(msg string) Command {
	response := chatServRespRegex.FindAllStringSubmatch(msg, -1)
	if (len(response) == 1) {
		return Command {
			Cmd: util.Decode(response[0][1]),
			User: util.Decode(response[0][2]),
			Body: util.Decode(response[0][3]),
		}
	} else {
		// If we have no match we must return an empty command.
		return Command{}
	}
}

// Grab the contents of the config file.
func getConfig() (string, util.Properties) {
	if(len(os.Args) >= 2) {
		user := os.Args[1]
		props := util.LoadConfig()
		return user, props
	} else {
		println("You must provide your username as the first argument!")
		os.Exit(1)
		return "", util.Properties{}
	}
}
