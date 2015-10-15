package main

import (
	"os"    // operating system functionality package
	"log"   // logging pkg
	"strings"
	"bufio" // buffered io
	"net"   // client/server pkg
	"fmt"   // formatted io
	"time"
)

const (
	CONN_PORT = ":1234"
	CONN_TYPE = "tcp"

	MAX_CLIENTS = 10

	CMD_PFX = "/"
	CMD_CREATE = CMD_PFX + "c"
	CMD_LIST   = CMD_PFX + "l"
	CMD_JOIN   = CMD_PFX + "j"
	CMD_LEAVE  = CMD_PFX + "leave"
	CMD_HELP   = CMD_PFX + "h"
	CMD_NAME   = CMD_PFX + "n"
	CMD_QUIT   = CMD_PFX + "q"

	CLIENT_NAME = "new_user" // TODO should implement new_user1, new_user2, etc
	SERVER_NAME = "Server"

	ERROR_PFX 		= "Error: "
	ERROR_SEND   	= ERROR_PFX + "You cannot send messages in the lobby.\n"
	ERROR_CREATE	= ERROR_PFX + "A chat room with that name already exists.\n"
	ERROR_JOIN   	= ERROR_PFX + "A chat room with that name does not exist.\n"
	ERROR_LEAVE  	= ERROR_PFX + "You cannot leave the lobby.\n"

	NOTICE_PFX          	= "Notice: "
	NOTICE_ROOM_JOIN       	= NOTICE_PFX + "\"%s\" joined.\n"
	NOTICE_ROOM_LEAVE      	= NOTICE_PFX + "\"%s\" left.\n"
	NOTICE_ROOM_NAME       	= NOTICE_PFX + "\"%s\" is now \"%s\".\n"
	NOTICE_ROOM_DELETE     	= NOTICE_PFX + "Inactive Room, Deleting...\n"
	NOTICE_LOBBY_CREATE 	= NOTICE_PFX + "Created \"%s\".\n"

	MSG_CONNECT = "Welcome. Type \"/h\" for commands.\n"
	MSG_FULL    = "Server is full."

	EXPIRY_TIME time.Duration = 7 * 24 * time.Hour
)


/* All users are placed in the lobby upon entry.
 * Allows /h commands to be used, but no messages otherwise
 * maps the list of recently (within a week) active chat rooms */
type Lobby struct {
	clients   []*Client
	chatRooms map[string]*ChatRoom
	incoming  chan *Message
	join      chan *Client
	leave     chan *Client
	delete    chan *ChatRoom
}

// Name of the chatroom, current clients, messagse, and expiry date and time. 
type ChatRoom struct {
	name     string
	clients  []*Client
	messages []string
	expiry   time.Time
}

// contains the clients name, current room, and connection info 
type Client struct {
	name     string
	chatRoom *ChatRoom
	incoming chan *Message
	outgoing chan string
	conn     net.Conn
	reader   *bufio.Reader
	writer   *bufio.Writer
}

// Contains the name of the sender, time, and text of a message
type Message struct {
	time   time.Time
	client *Client
	text   string
}

// create lobby
func NewLobby() *Lobby {
	lobby := &Lobby{
		clients:   make([]*Client, 0),
		chatRooms: make(map[string]*ChatRoom),
		incoming:  make(chan *Message),
		join:      make(chan *Client),
		leave:     make(chan *Client),
		delete:    make(chan *ChatRoom),
	}
	lobby.Listen()
	return lobby
}

// new lobby thread, listens for messages
func (lobby *Lobby) Listen() {
	go func() {
		for {
			select {
			case message := <-lobby.incoming:
				lobby.Parse(message)
			case client := <-lobby.join:
				lobby.Join(client)
			case client := <-lobby.leave:
				lobby.Leave(client)
			case chatRoom := <-lobby.delete:
				lobby.DeleteChatRoom(chatRoom)
			}
		}
	}()
}

// handles lobby connections
func (lobby *Lobby) Join(client *Client) {
	if len(lobby.clients) >= MAX_CLIENTS {
		client.Quit()
		return
	}
	lobby.clients = append(lobby.clients, client)
	client.outgoing <- MSG_CONNECT
	go func() {
		for message := range client.incoming {
			lobby.incoming <- message
		}
		lobby.leave <- client
	}()
}

// handles lobby disconnections
func (lobby *Lobby) Leave(client *Client) {
	if client.chatRoom != nil {
		client.chatRoom.Leave(client)
	}
	for i, otherClient := range lobby.clients {
		if client == otherClient {
			lobby.clients = append(lobby.clients[:i], lobby.clients[i+1:]...)
			break
		}
	}
	close(client.outgoing)
	log.Println("Closed client's outgoing channel")
}

// checks if channel is expired, deletes if so, sets new expiry time otherwise 
func (lobby *Lobby) DeleteChatRoom(chatRoom *ChatRoom) {
	if chatRoom.expiry.After(time.Now()) {
		go func() {
			time.Sleep(chatRoom.expiry.Sub(time.Now()))
			lobby.delete <- chatRoom
		}()
		log.Println("attempted to delete chat room")
	} else {
		chatRoom.Delete()
		delete(lobby.chatRooms, chatRoom.name)
		log.Println("deleted chat room")
	}
}

// creates a chatroom, unless that name is already in use
func (lobby *Lobby) CreateChatRoom(client *Client, name string) {
	if lobby.chatRooms[name] != nil {
		client.outgoing <- ERROR_CREATE
		log.Println("client tried to create chat room with a name already in use")
		return
	}
	chatRoom := NewChatRoom(name)
	lobby.chatRooms[name] = chatRoom
	go func() {
		time.Sleep(EXPIRY_TIME)
		lobby.delete <- chatRoom
	}()
	client.outgoing <- fmt.Sprintf(NOTICE_LOBBY_CREATE, chatRoom.name)
	log.Println("client created chat room")
}

/* joins a chat room (if it exists, warning otherwise). 
 * could have it create that room if it didnt exist and then join,
 * but then people could join rooms by mistake that are likely empty
 * example /j genral instead of /j general */
func (lobby *Lobby) JoinChatRoom(client *Client, name string) {
	if lobby.chatRooms[name] == nil {
		client.outgoing <- ERROR_JOIN
		log.Println("client tried to join a chat room that does not exist")
		return
	}
	if client.chatRoom != nil {
		lobby.LeaveChatRoom(client)
	}
	lobby.chatRooms[name].Join(client)
	log.Println("client joined chat room")
}

// leaves chat room
func (lobby *Lobby) LeaveChatRoom(client *Client) {
	if client.chatRoom == nil {
		client.outgoing <- ERROR_LEAVE
		log.Println("client tried to leave the lobby")
		return
	}
	client.chatRoom.Leave(client)
	log.Println("client left chat room")
}

// lists currently open chat rooms
func (lobby *Lobby) ListChatRooms(client *Client) {
	client.outgoing <- "\n"
	client.outgoing <- "Chat Rooms:\n"
	for name := range lobby.chatRooms {
		client.outgoing <- fmt.Sprintf("%s\n", name)
	}
	client.outgoing <- "\n"
	log.Println("client listed chat rooms")
}

// creates a new chat room, sets expiration date
func NewChatRoom(name string) *ChatRoom {
	return &ChatRoom{
		name:     name,
		clients:  make([]*Client, 0),
		messages: make([]string, 0),
		expiry:   time.Now().Add(EXPIRY_TIME),
	}
}


// checks for prefix commands first, otherwise sends a message 
func (lobby *Lobby) Parse(message *Message) {
	switch {
	default:
		lobby.SendMessage(message)
	case strings.HasPrefix(message.text, CMD_CREATE):
		name := strings.TrimSuffix(strings.TrimPrefix(message.text, CMD_CREATE+" "), "\n")
		lobby.CreateChatRoom(message.client, name)
	case strings.HasPrefix(message.text, CMD_LEAVE):
		lobby.LeaveChatRoom(message.client)
	case strings.HasPrefix(message.text, CMD_LIST):
		lobby.ListChatRooms(message.client)
	case strings.HasPrefix(message.text, CMD_JOIN):
		name := strings.TrimSuffix(strings.TrimPrefix(message.text, CMD_JOIN+" "), "\n")
		lobby.JoinChatRoom(message.client, name)
	case strings.HasPrefix(message.text, CMD_NAME):
		name := strings.TrimSuffix(strings.TrimPrefix(message.text, CMD_NAME+" "), "\n")
		lobby.ChangeName(message.client, name)
	case strings.HasPrefix(message.text, CMD_HELP):
		lobby.Help(message.client)
	case strings.HasPrefix(message.text, CMD_QUIT):
		message.client.Quit()
	}
}

// sends message to chat room. error message if in the lobbby
func (lobby *Lobby) SendMessage(message *Message) {
	if message.client.chatRoom == nil {
		message.client.outgoing <- ERROR_SEND
		log.Println("client tried to send message in lobby")
		return
	}
	message.client.chatRoom.Broadcast(message.String())
	log.Println("client sent message")
}

// change user name
func (lobby *Lobby) ChangeName(client *Client, name string) {
	if client.chatRoom == nil {
		client.outgoing <- (fmt.Sprintf(NOTICE_ROOM_NAME, client.name, name))
	} else {
		client.chatRoom.Broadcast(fmt.Sprintf(NOTICE_ROOM_NAME, client.name, name))
	}
	client.name = name
	log.Println("client changed their name")
}

// sends list of commands
func (lobby *Lobby) Help(client *Client) {
	client.outgoing <- "\n"
	client.outgoing <- "Commands:\n"
	client.outgoing <- CMD_HELP +" - lists all commands\n"
	client.outgoing <- CMD_LIST + " - lists all chat rooms\n"
	client.outgoing <- CMD_CREATE + " test - creates a chat room named test\n"
	client.outgoing <- CMD_JOIN + " test - joins a chat room named test\n"
	client.outgoing <- CMD_LEAVE + " - leaves the current chat room\n"
	client.outgoing <- CMD_NAME + " test - changes your name to test\n"
	client.outgoing <- CMD_QUIT + " - quits the program\n"
	client.outgoing <- "\n"
	log.Println("client requested help")
}

// sends all of the previous message upon joining the chat room
func (chatRoom *ChatRoom) Join(client *Client) {
	client.chatRoom = chatRoom
	for _, message := range chatRoom.messages {
		client.outgoing <- message
	}
	chatRoom.clients = append(chatRoom.clients, client)
	chatRoom.Broadcast(fmt.Sprintf(NOTICE_ROOM_JOIN, client.name))
}

// Removes client from chat room.
func (chatRoom *ChatRoom) Leave(client *Client) {
	chatRoom.Broadcast(fmt.Sprintf(NOTICE_ROOM_LEAVE, client.name))
	for i, otherClient := range chatRoom.clients {
		if client == otherClient {
			chatRoom.clients = append(chatRoom.clients[:i], chatRoom.clients[i+1:]...)
			break
		}
	}
	client.chatRoom = nil
}

// sends the current chatroom the message
func (chatRoom *ChatRoom) Broadcast(message string) {
	chatRoom.expiry = time.Now().Add(EXPIRY_TIME)
	chatRoom.messages = append(chatRoom.messages, message)
	for _, client := range chatRoom.clients {
		client.outgoing <- message
	}
}

// Notifies the clients within the chat room that it is being deleted, and kicks
// them back into the lobby.
func (chatRoom *ChatRoom) Delete() {
	//notify of deletion?
	chatRoom.Broadcast(NOTICE_ROOM_DELETE)
	for _, client := range chatRoom.clients {
		client.chatRoom = nil
	}
}



// creates a new client, opens reader/writer for them.
func NewClient(conn net.Conn) *Client {
	writer := bufio.NewWriter(conn)
	reader := bufio.NewReader(conn)

	client := &Client{
		name:     CLIENT_NAME,
		chatRoom: nil,
		incoming: make(chan *Message),
		outgoing: make(chan string),
		conn:     conn,
		reader:   reader,
		writer:   writer,
	}

	client.Listen()
	return client
}

// creates two threads to read and write
func (client *Client) Listen() {
	go client.Read()
	go client.Write()
}

/* reads string from client, formats into message or returns error. 
 * sends it back to client */
func (client *Client) Read() {
	for {
		str, err := client.reader.ReadString('\n')
		if err != nil {
			log.Println(err)
			break
		}
		message := NewMessage(time.Now(), client, strings.TrimSuffix(str, "\n"))
		client.incoming <- message
	}
	close(client.incoming)
	log.Println("Closed client's incoming channel read thread")
}

// reads message from outgoing, writes to socket
func (client *Client) Write() {
	for str := range client.outgoing {
		_, err := client.writer.WriteString(str)
		if err != nil {
			log.Println(err)
			break
		}
		err = client.writer.Flush()
		if err != nil {
			log.Println(err)
			break
		}
	}
	log.Println("Closed client's write thread")
}

// close clients connection
func (client *Client) Quit() {
	client.conn.Close()
}


// Creates a new message with the given time, client and text.
func NewMessage(time time.Time, client *Client, text string) *Message {
	return &Message{
		time:   time,
		client: client,
		text:   text,
	}
}

// returns a string with time, sender, and message (from NewMessage)
func (message *Message) String() string {
	return fmt.Sprintf("%s - %s: %s\n", message.time.Format(time.Kitchen), message.client.name, message.text)
}

// creates the lobby, listens for connections
func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	lobby := NewLobby()

	listener, err := net.Listen(CONN_TYPE, CONN_PORT)
	if err != nil {
		log.Println("Error: ", err)
		os.Exit(1)
	}
	defer listener.Close()
	log.Println("Listening on " + CONN_PORT)

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Println("Error: ", err)
			continue
		}
		lobby.Join(NewClient(conn))
	}
}