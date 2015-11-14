package main

import (
	"log"
	"net"
	"net/http"
	"net/rpc"
	"time"

)

const (
	TYPE = "tcp"
	PORT = ":65535"
	HOST = "localhost"
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
	ERR_TOK	   = ERR_PREFIX + "User not found with this token!\n"

	/*Client name, followed by the server name.*/
	CNAME = "Anon"
	SNAME = "MyServer"

	/*Maximum amount of clients allowed to connect.*/
	CLIENTMAX = 10

	/*Server specific messages, such as a welcome and reason as to why a user
	* cannot connect (server being full)*/
	SFULL    = "Sorry, the server is currently full! Please, try again later.\n"
	SCONNECT = "Welcome to the GO Chat server! To get started, type \"!help\" to retrieve a list of commands.\n"

	/*An expiry time for messages, they have to be seven days old to be deleted.*/
	EXTIME time.Duration = 7 * 24 * time.Hour

)

func main() {
	  receive := new(Receiver)
	  rpc.Register(receive)
	  rpc.HandleHTTP()
	  listen, err := net.Listen(TYPE, PORT)
	  if err != nil {
		log.Fatal("Listen error: ", err)
	  
	  }
	  log.Print("Starting server")
	  http.Serve(listen, nil)

}