/* File to handle the server input, and user input to reflect back to other
* clients.*/
package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"os"
	"strings"
	"time"
)

/*Create the constants we will use in this file, so we can interchange them
* easily.*/

const (
	TYPE = "tcp"
	PORT = ":65535"
	HOST = "localhost"
	/*Prefix, and following commands for that chat rooms.*/
	COMM_PREFIX     = "!"
	COMM_CREATEROOM = COMM_PREFIX + "create"
	COMM_ENTERROOM  = COMM_PREFIX + "enter"
	COMM_LEAVEROOM  = COMM_PREFIX + "leave"
	COMM_LISTROOMS  = COMM_PREFIX + "list"

	COMM_CHANGENAME = COMM_PREFIX + "name"
	COMM_QUITCHAT   = COMM_PREFIX + "quit"
	COMM_HELPCHAT   = COMM_PREFIX + "help"

	/*Notices for the server to output whenever a user joins, etc.*/
	NOTE_PREFIX         = "Note: "
	NOTE_CHANGENAME     = NOTE_PREFIX + "Changed their name to [%s].\n"
	NOTE_ROOM_CREATE    = NOTE_PREFIX + "Created the room {%s}.\n"
	NOTE_ROOM_ENTER     = NOTE_PREFIX + "[%s] has joined the room.\n"
	NOTE_ROOM_LEAVE     = NOTE_PREFIX + "[%s] has left the room.\n"
	NOTE_PUB_CHANGENAME = NOTE_PREFIX + "[%s] changed their name to [%s].\n"
	NOTE_ROOM_DELETION  = NOTE_PREFIX + "Chat room is being deleted due to inactivity.\n"

	/*List of error commands that a user can encounter.*/
	ERR_PREFIX = "Error: "
	ERR_CREATE = ERR_PREFIX + "There is a chat room with that name already.\n"
	ERR_ENTER  = ERR_PREFIX + "Chat room does not exist, you cannot join.\n"
	ERR_LEAVE  = ERR_PREFIX + "You cannot leave the lobby!\n"
	ERR_SEND   = ERR_PREFIX + "Cannot send messages in the lobby.\n"

	/*Client name, followed by the server name.*/
	CNAME = "Anon"
	SNAME = "MyServer"

	/*Maximum amount of clients allowed to connect.*/
	CLIENTMAX = 12

	/*Server specific messages, such as a welcome and reason as to why a user
	* cannot connect (server being full)*/
	SFULL    = "Sorry, the server is currently full! Please, try again later.\n"
	SCONNECT = "Welcome to the GO Chat server! To get started, type \"!help\" to retrieve a list of commands.\n"

	/*An expiry time for messages, they have to be seven days old to be deleted.*/
	EXTIME time.Duration = 7 * 24 * time.Hour
)

/*Begin with client features, such as adding a new client for a reader writer
*and quitting.*/
type Client struct {
	connect    net.Conn
	readWatch  *bufio.Reader
	writeWatch *bufio.Writer
	incMsg     chan *Msg
	outMsg     chan string
	cRoom      *CRoom
	username   string
}

/*Create a constructor for a client, which will set the client to a deafult name
* and will set the socket of the connection by the provided input.*/
func ClientConstruct(connect net.Conn) *Client {
	writeWatch := bufio.NewWriter(connect)
	readWatch := bufio.NewReader(connect)

	newClient := &Client{
		connect:    connect,
		readWatch:  readWatch,
		writeWatch: writeWatch,
		incMsg:     make(chan *Msg),
		outMsg:     make(chan string),
		cRoom:      nil,
		username:   CNAME,
	}
	newClient.Listen()
	return newClient
}

/*Need to create a listen to the client function, which will just start a reader
* and a writer for the client.*/
func (client *Client) Listen() {
	go client.ReadMsg()
	go client.WriteMsg()
}

/*This function will read message from the clients outMsg, and write them to
* the clients socket.*/
func (client *Client) WriteMsg() {
	/*As long as the message is as long as the outgoing channel, write the
	* string.*/
	for msg := range client.outMsg {
		_, errNo := client.writeWatch.WriteString(msg)
		if errNo != nil {
			log.Println(errNo)
			break
		}
		/*Make sure we error check the flush, because it could fail.*/
		errNo = client.writeWatch.Flush()

		if errNo != nil {
			log.Println(errNo)
			break
		}
	}

	log.Println("Write thread is now closed for client.")
}

/*Close the client connection if they wish to quit.*/
func (client *Client) Quit() {
	client.connect.Close()
}

/*Function will act as the reader for the function to handle input from client
* and format into messages and put them in the client channel.*/
func (client *Client) ReadMsg() {
	for {
		incMsg, errNo := client.readWatch.ReadString('\n')
		if errNo != nil {
			log.Println(errNo)
			break
		}
		msg := NewMsg(time.Now(), client, strings.TrimSuffix(incMsg, "\n"))
		client.incMsg <- msg
	}
	close(client.incMsg)
	log.Println("Closed read channel of client thread.")
}

/*Lobby stuff next, as the client is initially connected to the lobby, and
*the lobby has information about all the rooms and everything.*/
type Lobby struct {
	curClients []*Client
	cRoom      map[string]*CRoom
	incMsg     chan *Msg
	delRoom    chan *CRoom
	joinRoom   chan *Client
	leaveRoom  chan *Client
}

/*Create a new lobby which listens over all channels.*/
func NewLobby() *Lobby {
	newLob := &Lobby{
		curClients: make([]*Client, 0),
		cRoom:      make(map[string]*CRoom),
		incMsg:     make(chan *Msg),
		delRoom:    make(chan *CRoom),
		joinRoom:   make(chan *Client),
		leaveRoom:  make(chan *Client),
	}
	newLob.Listen()
	return newLob
}

/*Start a new lobby thread which listen over all of lobby's channels.*/
func (lob *Lobby) Listen() {
	go func() {
		for {
			select {
			case msg := <-lob.incMsg:
				lob.ParseMsg(msg)
			case client := <-lob.joinRoom:
				lob.JoinRoom(client)
			case client := <-lob.leaveRoom:
				lob.LeaveCRoom(client)
			case cRoom := <-lob.delRoom:
				lob.DeleteCRoom(cRoom)
			}
		}
	}()
}

/*JoinRoom handles the new clients that connect to the lobby, mainly the
*initial server connection is handled here.*/
func (lob *Lobby) JoinRoom(client *Client) {
	/*If the clients has reached the max amount of people, do not allow the
	* client to join.*/
	if len(lob.curClients) >= CLIENTMAX {
		client.Quit()
		return
	}
	/*Add user to array of users.*/
	lob.curClients = append(lob.curClients, client)
	/*Add the welcome message to the clients structure.*/
	client.outMsg <- SCONNECT
	/*Add the latest incoming message to the lobby, for each client.*/
	go func() {
		for msg := range client.incMsg {
			lob.incMsg <- msg
		}
		lob.leaveRoom <- client
	}()
}

/*This function will handle leaving a room.*/
func (lob *Lobby) Leave(client *Client) {
	/*If the client is in a room, leave.*/
	if client.cRoom != nil {
		client.cRoom.Leave(client)
	}
	/*Add the user back to the lobby clients.*/
	for k, oClient := range lob.curClients {
		if client == oClient {
			lob.curClients = append(lob.curClients[:k], lob.curClients[k+1:]...)
			break
		}
	}
	close(client.outMsg)
	log.Println("Closed the outgoing channel for the client.")
}

/*Delete will check if a certain channel has expired, if so the room will be
* deleted. Otherwise, a signal will be sent to the channel to be deleted
* for its new expiry time.*/
func (lob *Lobby) DeleteCRoom(cRoom *CRoom) {
	if cRoom.expire.After(time.Now()) {
		go func() {
			time.Sleep(cRoom.expire.Sub(time.Now()))
			lob.delRoom <- cRoom
		}()
		log.Println("Attempted to delete a chat room.")
	} else {
		cRoom.Delete()
		delete(lob.cRoom, cRoom.cName)
		log.Println("Deleted the room successfully.")
	}
}

/*This function will remove a user from the current chat room, and will check
* if they are already in a room.*/
func (lob *Lobby) LeaveCRoom(client *Client) {
	if client.cRoom == nil {
		client.outMsg <- ERR_LEAVE
		log.Println("Error in making a user leave a room.\n")
		return
	}
	client.cRoom.Leave(client)
	log.Println("Succesfully made user leave a chat room!\n")
}

/*We need a function to join chat. Will try and join a user to a room,
* provided it exists.*/
func (lob *Lobby) EnterCRoom(client *Client, cRoomName string) {
	/*Check if room exists.*/
	if lob.cRoom[cRoomName] == nil {
		client.outMsg <- ERR_ENTER
		log.Println("Attempted to add user to a room that doesn't exist.\n")
		return
	}
	/*Check if user doesn't have a room in their current state.*/
	/*If so, just leave their old chat room to join this one.*/
	if client.cRoom != nil {
		lob.LeaveCRoom(client)
	}
	lob.cRoom[cRoomName].Join(client)
	log.Println("Successfully added user to another room.")
}

/*The next two functions will deal with creating and deleting chatrooms.
* Create will attempt to create a room with a user-specfied name, if it
* exists it will not let it, as the name is already in use. We have this
* function below entercroom so it can reference it and drop the user into
* the room upon creation.*/
func (lob *Lobby) CreateCRoom(client *Client, cRoomName string) {
	/*Check if the name already exists.*/
	if lob.cRoom[cRoomName] != nil {
		client.outMsg <- ERR_CREATE
		log.Println("User tried to create a room with a name that is already in use.\n")
		return
	}
	/*Create a new chat room, add it to the chat room lobby channel.*/
	cRoom := NewCRoom(cRoomName)
	lob.cRoom[cRoomName] = cRoom
	go func() {
		time.Sleep(EXTIME)
		lob.delRoom <- cRoom
	}()
	client.outMsg <- fmt.Sprintf(NOTE_ROOM_CREATE, cRoom.cName)
	lob.EnterCRoom(client, cRoomName)
	log.Println("User created a new chat room!\n")
}

/*Begin some user commands, such as messaging, changing names, listing rooms
* etc.*/
/*Broadcast a message to the entire channel the user is in.*/
func (lob *Lobby) SendMsg(msg *Msg) {
	if msg.client.cRoom == nil {
		msg.client.outMsg <- ERR_SEND
		log.Println("Client tried to send a message in the lobby.")
		return
	}
	msg.client.cRoom.Broadcast(msg.toString())
	log.Println("Sucess on sending client message.")
}

/*Changes the clients name to a given different name.*/
func (lob *Lobby) ChangeUsername(client *Client, username string) {
	if client.cRoom == nil {
		client.outMsg <- fmt.Sprintf(NOTE_CHANGENAME, username)
	} else {
		client.cRoom.Broadcast(fmt.Sprintf(NOTE_PUB_CHANGENAME, client.username, username))
	}
	client.username = username
	log.Println("Success on client changing name!\n")
}

/*List all the current chat rooms that are active to the user.*/
func (lob *Lobby) ListCRooms(client *Client) {
	/*Throw in a new line for relob.Help(msg.client)factor purposes.*/
	client.outMsg <- "\n\n"
	client.outMsg <- "Chat Rooms:\n"
	/*Print out all the rooms, through a for loop.*/
	for cName := range lob.cRoom {
		client.outMsg <- fmt.Sprintf("%s\n", cName)
	}
	client.outMsg <- "\n"
	log.Println("Client sucess on listing chat rooms.\n")
}

/*The command for help, will just have several messages sent to the channel.
* Each will tell the command and what it does.*/
func (lob *Lobby) Help(client *Client) {
	/*Refactoring purpose.*/
	client.outMsg <- "\n\n"
	client.outMsg <- "Commands and Usage:\n"
	client.outMsg <- "!help - lists all commands.\n"
	client.outMsg <- "!list - lists all chat rooms that are active.\n"
	client.outMsg <- "!name param - changes your name to param.\n"
	client.outMsg <- "!create chan - creates a channel called chan.\n"
	client.outMsg <- "!enter chan - enters a chat named chan.\n"
	client.outMsg <- "!leave - leaves the current channel.\n"
	client.outMsg <- "!quit - quits the chat client.\n"
	client.outMsg <- "\n\n"
	log.Println("User accessed the help section.\n")
}

/*Extra command for parsing. Handles all messages sent to lobby, will check
* for the command prefix, if not it will just send a message.*/
func (lob *Lobby) ParseMsg(msg *Msg) {
	switch {
	case strings.HasPrefix(msg.txt, COMM_CREATEROOM):
		cName := strings.TrimSuffix(strings.TrimPrefix(msg.txt, COMM_CREATEROOM+" "), "\n")
		lob.CreateCRoom(msg.client, cName)
	case strings.HasPrefix(msg.txt, COMM_ENTERROOM):
		cName := strings.TrimSuffix(strings.TrimPrefix(msg.txt, COMM_ENTERROOM+" "), "\n")
		lob.EnterCRoom(msg.client, cName)
	case strings.HasPrefix(msg.txt, COMM_LEAVEROOM):
		lob.LeaveCRoom(msg.client)
	case strings.HasPrefix(msg.txt, COMM_LISTROOMS):
		lob.ListCRooms(msg.client)
	case strings.HasPrefix(msg.txt, COMM_HELPCHAT):
		lob.Help(msg.client)
	case strings.HasPrefix(msg.txt, COMM_CHANGENAME):
		username := strings.TrimSuffix(strings.TrimPrefix(msg.txt, COMM_CHANGENAME+" "), "\n")
		lob.ChangeUsername(msg.client, username)
	case strings.HasPrefix(msg.txt, COMM_QUITCHAT):
		msg.client.Quit()
	default:
		lob.SendMsg(msg)
	}
}

/*Chat rooms are a part of the lobby, which contains the chat room name, list
* of all connected clients to that channel, and a history of the messages
* that were sent, and a time in which the room will expire.*/
type CRoom struct {
	cName      string
	curClients []*Client
	msgs       []string
	expire     time.Time
}

/*Creation of a new chat room, simply return a room with a given string.*/
func NewCRoom(cName string) *CRoom {
	return &CRoom{
		cName:      cName,
		curClients: make([]*Client, 0),
		msgs:       make([]string, 0),
		expire:     time.Now().Add(EXTIME),
	}
}

/*Start off with a broadcast command, which will just send a message to the
* outmsg of a client, to each user within it's channel.*/
func (cRoom *CRoom) Broadcast(msg string) {
	/*Rooms been accessed, increase the time of expiry.*/
	cRoom.expire = time.Now().Add(EXTIME)
	cRoom.msgs = append(cRoom.msgs, msg)
	for _, client := range cRoom.curClients {
		client.outMsg <- msg
	}
}

/*Now we need to make functions for creation, deletion and joining of different
* chat rooms. Delete will simply notify that the channel is going to be
* deleted due to inactivity.*/
func (cRoom *CRoom) Delete() {
	/*If there's people in the room, notify of deletion.*/
	/*Basically need to rewrite function on inside, or it will renew the
	* room.*/
	//cRoom.Broadcast(NOTE_ROOM_DELETION)
	for _, client := range cRoom.curClients {
		client.cRoom = nil
	}
	//cRoom.curClients = nil
	//cRoom.msgs = nil
	//cRoom.expire = nil
}

/*When a user joins a room, we want to notify the people that he has joined.
* Also, set his room to his chat room, and give him the history of messages
* as he joins.*/
func (cRoom *CRoom) Join(client *Client) {
	client.cRoom = cRoom
	if len(cRoom.msgs) != 0 {
		client.outMsg <- "================BEGIN LOG================\n"
	}

	for _, msg := range cRoom.msgs {
		client.outMsg <- msg
	}
	if len(cRoom.msgs) != 0 {
		client.outMsg <- "================END LOG================\n"
	}
	cRoom.curClients = append(cRoom.curClients, client)
	cRoom.Broadcast(fmt.Sprintf(NOTE_ROOM_ENTER, client.username))
}

/*Delete will require to remove a user from the room as well, uses a similar
* for loop as seen from above. Notify the user is leaving, remove him, and
* set his place back to lobby.*/
func (cRoom *CRoom) Leave(client *Client) {
	cRoom.Broadcast(fmt.Sprintf(NOTE_ROOM_LEAVE, client.username))
	for k, oClient := range cRoom.curClients {
		if client == oClient {
			cRoom.curClients = append(cRoom.curClients[:k], cRoom.curClients[k+1:]...)
			break
		}
	}
	client.cRoom = nil
}

/*Messages are the structure that is a part of the chat room, which chat rooms
* are a part of the lobby. They contain a string message, the timestamp of when
* the message was sent, and the client themselves, so we can track who
* said what.*/
type Msg struct {
	time   time.Time
	client *Client
	txt    string
}

/*Create a new message with given user, text and timestamp.*/
func NewMsg(time time.Time, client *Client, txt string) *Msg {
	return &Msg{
		time:   time,
		client: client,
		txt:    txt,
	}
}

/*Now we need to return a string representation of the message, as sending
* the entire structure will not work.*/
func (msg *Msg) toString() string {
	return fmt.Sprintf("%s-[%s]  %s\n", msg.time.Format(time.Kitchen), msg.client.username, msg.txt)
}

/*Main will create a single lobby, listen for user and connect them to a lobby
* and await commands from the user.*/
func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	lob := NewLobby()

	listen, errNo := net.Listen(TYPE, HOST+PORT)
	if errNo != nil {
		log.Println("Error: ", errNo)
		os.Exit(1)
	}
	defer listen.Close()
	log.Println("Listening on server " + HOST + PORT)

	for {
		connect, errNo := listen.Accept()
		if errNo != nil {
			log.Println("Error: ", errNo)
			continue
		}
		lob.JoinRoom(ClientConstruct(connect))
	}
}
