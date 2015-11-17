package main

import (
	"errors"
	"fmt"
	"log"
	"sync"
	"time"
  
)

/*Chat rooms are a part of the lobby, which contains the chat room name, list
* of all connected clients to that channel, and a history of the messages
* that were sent, and a time in which the room will expire.*/
type CRoom struct {
	cName      string
	curClients []*Client
	msgs       []string
	join       chan *Client
	leave      chan *Client
	incoming   chan string
	expiry     chan bool
	expire     time.Time
}

var cRooms map[string]*CRoom = make(map[string]*CRoom)
var cRoomMutex sync.RWMutex

/*Creation of a new chat room, simply return a room with a given string.*/
func NewCRoom(cName string) *CRoom {
	cRoom := &CRoom{
		cName:      cName,
		curClients: make([]*Client, 0),
		msgs:       make([]string, 0),
		join:	    make(chan *Client),
		leave:      make(chan *Client),
		incoming:   make(chan string),
		expiry:     make(chan bool),
		expire:     time.Now().Add(EXTIME),
	}
	cRoom.Listen()
	cRoom.Delete()
	return cRoom
}

func (chatRoom *CRoom) Listen() {
	go func() {
		for {
			select {
			case client := <-chatRoom.join:
				chatRoom.AddClient(client)
			case client := <-chatRoom.leave:
				chatRoom.removeClient(client)
			case message := <-chatRoom.incoming:
				chatRoom.Broadcast(message)
			case _ = <-chatRoom.expiry:
				chatRoom.Delete()
			}
		}
	}()
}

func (cRoom *CRoom) Delete() {
  log.Println("Deleting.")
  if cRoom.expire.After(time.Now()) {
    go func () {
      time.Sleep(cRoom.expire.Sub(time.Now()))
      cRoom.expiry <- true
    }()
  } else {
    cRoom.Broadcast(NOTE_ROOM_DELETION)
    for _, client := range cRoom.curClients {
      client.Mutex.Lock()
      client.CRoom = nil
      client.Mutex.Unlock()
    }
    deleteCRoom(cRoom.cName)
  }
}

func (cRoom *CRoom) AddClient(client *Client) {
  log.Println("Adding a new client")
  client.Mutex.Lock()
  defer client.Mutex.Unlock()
  
  cRoom.Broadcast(fmt.Sprintf(NOTE_ROOM_ENTER, client.Name))
  for _, msg := range cRoom.msgs {
    client.outMsg <- msg
  }
  cRoom.curClients = append(cRoom.curClients, client)
  client.CRoom = cRoom
}


func (cRoom *CRoom) removeClient(client *Client) {
  client.Mutex.RLock()
  cRoom.Broadcast(fmt.Sprintf(NOTE_ROOM_LEAVE, client.Name))
  client.Mutex.RUnlock()
  for i, oClients := range cRoom.curClients {
    if client == oClients {
      cRoom.curClients = append(cRoom.curClients[:i], cRoom.curClients[i+1:]...)
      break
    }
  }
  client.Mutex.Lock()
  defer client.Mutex.Unlock()
  client.CRoom = nil
  log.Println("Removed client\n")
}

func (cRoom *CRoom) Broadcast(msg string) {
  cRoom.expire = time.Now().Add(EXTIME)
  log.Println(msg)
  cRoom.msgs = append(cRoom.msgs, msg)
  for _, client := range cRoom.curClients {
    client.outMsg <- msg
  }
}

func addCRoom (cRoom *CRoom) error {
  cRoomMutex.Lock()
  defer cRoomMutex.Unlock()
  
  oCRoom := cRooms[cRoom.cName]
  if oCRoom != nil {
    return errors.New(ERR_CREATE)
  }
  cRooms[cRoom.cName] = cRoom
  return nil
}

func deleteCRoom (cName string) error {
  cRoomMutex.Lock()
  defer cRoomMutex.Unlock()
  
  cRoom := cRooms[cName]
  if cRoom == nil {
    return errors.New(ERR_ENTER)
  }
  delete(cRooms, cName)
  return nil
}

func getCRoom (cName string) (*CRoom, error) {
    cRoomMutex.RLock()
    defer cRoomMutex.RUnlock()
    
    cRoom := cRooms[cName]
    if cRoom == nil {
      return nil, errors.New(ERR_ENTER)
    }
    return cRoom, nil
}

func getCRoomName() []string {
  cRoomMutex.RLock()
  defer cRoomMutex.RUnlock()
  
  keys := make([]string, 0, len(cRooms))
  for k := range cRooms {
    keys = append(keys, k)
  }
  return keys
}

