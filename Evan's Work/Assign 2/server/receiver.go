package main

import (
  "crypto/rand"
  "encoding/base64"
  "fmt"
  "log"
  "time"
)

type Args struct {
  Token string
  String string
}

type Receiver int

func randString(len int) (str string) {
  by := make([]byte, len)
  rand.Read(by)
  return base64.StdEncoding.EncodeToString(by)
}

func (r *Receiver) Connect(args *struct{}, tok *string) error {
  log.Println("Connecting.\n")
  *tok = randString(64)
  client := NewClient(*tok)
  err := AddClient(client)
  if err != nil {
    log.Println(err)
    return err
  }
  go func() {client.outMsg <- SCONNECT }()
  return nil
}

func (r *Receiver) RecMsg(tok *string, msg *string) error {
  log.Println("recMsg")
  client, err := GetClient(*tok)
  if err != nil {
    return err
  }
  *msg = <-client.outMsg
  return nil

}

func (r *Receiver) SendMsg(args *Args, _ *struct{}) error {
  log.Println("Sending a message.\n")
  client, err := GetClient(args.Token)
  if err != nil {
    log.Println(err)
    return err
  }
  client.Mutex.RLock()
  defer client.Mutex.RUnlock()
  if client.CRoom == nil {
    client.outMsg <- ERR_SEND
    return nil
  }
  client.CRoom.incoming <- fmt.Sprintf("[%s] - %s: %s", time.Now().Format(time.Kitchen), client.Name, args.String)
  return nil
}

func (r *Receiver) Quit(tok *string, _ *struct{}) error {
  log.Println("User quitting\n")
  err := RemoveClient(*tok)
  if err != nil {
    return err
  }
  return nil
}

/*Functions for the chat rooms for reciever.*/
func (r *Receiver) CreateCRoom(args *Args, _ *struct{}) error {
  log.Println("Creating a chatroom now.\n")
  client, err := GetClient(args.Token)
  if err != nil {
    log.Println(err)
    return err
  }
  cRoom := NewCRoom(args.String)
  err = addCRoom(cRoom)
  if err != nil {
    client.outMsg <- err.Error()
    log.Println(err)
    return err
  }
  client.outMsg <- fmt.Sprintf(NOTE_ROOM_CREATE)
  return nil
}

func (r *Receiver) JoinCRoom (args *Args, _ *struct{}) error {
  log.Println("Joining chat Room now.\n")
  client, err := GetClient(args.Token)
  if err != nil {
    log.Println(err)
    return err  
  }
  cRoom, err := getCRoom(args.String)
  if err != nil {
    client.outMsg <- err.Error()
    log.Println(err)
    return err
  }
  client.Mutex.RLock()
  oldCRoom := client.CRoom
  client.Mutex.RUnlock()
  if oldCRoom != nil {
    oldCRoom.leave <- client
  }
  cRoom.join <- client
  return nil
}

func (r *Receiver) LeaveCRoom(tok *string, _ *struct{}) error {
  log.Println("Leaving a Chat Room\n")
  client, err := GetClient(*tok)
  if err != nil {
    log.Println(err)
    return err
  }
  if client.CRoom == nil {
    log.Println("User tried to leave lobby")
    return err
  }
  client.Mutex.RLock()
  defer client.Mutex.RUnlock()
  log.Println(client)
  client.CRoom.leave <- client
  return nil
}

func (r *Receiver) ListCRooms(tok *string, _ *struct{}) error {
  log.Println("Listing all chat rooms.\n")
  client, err := GetClient(*tok)
  if err != nil {
    log.Println(err)
    return err
  }
  cRoomNames := getCRoomName()
  cList := "\n=====Chat Rooms======\n"
  for _, cRoomName := range cRoomNames {
    cList += cRoomName + "\n"
  }
  cList += "\n"
  client.outMsg <- cList
  return nil
}

func (r *Receiver) ChangeName(args *Args, _ *struct{}) error {
  log.Println("Changing name now.\n")
  client, err := GetClient(args.Token)
  if err != nil {
    log.Println("Error changing null client name.\n")
    return err
  }
  client.Mutex.Lock()
  defer client.Mutex.Unlock()
  client.Name = args.String
  return nil
}