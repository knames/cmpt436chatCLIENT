package main

import (
  "bufio"
  "fmt"
  "net/rpc"
  "os"
  "strings"
  "log"
  "sync"

)

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
	
	/*Help messages refactored..*/
	MESS_HELP = "\nCommands and Usage:\n" +
	  COMM_HELPCHAT + " - lists all commands.\n" +
	  COMM_LISTROOMS + " - lists all chat rooms that are active.\n" +
	  COMM_CHANGENAME + " - changes your name to param.\n" +
	  COMM_CREATEROOM + " chan - creates a channel called chan.\n" +
	  COMM_ENTERROOM + " - enters a chat named chan.\n" +
	  COMM_LEAVEROOM + " - leaves the current channel.\n" +
	  COMM_QUITCHAT + " - quits the chat client.\n" 
	  
	MESS_DISC = "Goodbye, disconnect.\n"
)
type Args struct {
    Token string
    String string
}

var Token string
var client *rpc.Client
var waitG sync.WaitGroup

func UserIn() {
  reader := bufio.NewReader(os.Stdin)
  for {
      usrStr, err := reader.ReadString('\n')
      if err != nil {
	waitG.Done()
	break
      }
      
      usrParse(usrStr)
      
  }
  
}

func usrParse(usrStr string) (err error) {
    switch {
	default:
		err = client.Call("Receiver.SendMsg", Args{Token, usrStr}, nil)
	case strings.HasPrefix(usrStr, COMM_CREATEROOM):
		cName := strings.TrimSuffix(strings.TrimPrefix(usrStr, COMM_CREATEROOM+" "), "\n")
		err = client.Call("Receiver.CreateCRoom", Args{Token, cName}, nil)
	case strings.HasPrefix(usrStr, COMM_ENTERROOM):
		cName := strings.TrimSuffix(strings.TrimPrefix(usrStr, COMM_ENTERROOM+" "), "\n")
		err = client.Call("Receiver.JoinCRoom", Args{Token, cName}, nil)
	case strings.HasPrefix(usrStr, COMM_LEAVEROOM):
		err = client.Call("Receiver.LeaveCRoom", &Token, nil)
	case strings.HasPrefix(usrStr, COMM_LISTROOMS):
		cName := strings.TrimSuffix(strings.TrimPrefix(usrStr, COMM_LISTROOMS+" "), "\n")
		err = client.Call("Receiver.ListCRooms", Args{Token, cName}, nil)
	case strings.HasPrefix(usrStr, COMM_HELPCHAT):
		fmt.Print(MESS_HELP)
	case strings.HasPrefix(usrStr, COMM_CHANGENAME):
		cName := strings.TrimSuffix(strings.TrimPrefix(usrStr, COMM_CHANGENAME+" "), "\n")
		err = client.Call("Receiver.ChangeName", Args{Token, cName}, nil)
	case strings.HasPrefix(usrStr, COMM_QUITCHAT):
		err = client.Call("Receiver.Quit", &Token, nil)
		waitG.Done()
	}
	//fmt.Print(err)
	return err
}

func userOut() {
    for {
	var msg string
	err := client.Call("Receiver.RecMsg", &Token, &msg)
	//err := client.Call("Receiver.RecMsg", &Token, &msg)
	if err != nil {
	    fmt.Print(err)
	    waitG.Done()
	    break
	}
	fmt.Print(msg)
    }
    
}

func main() {
    waitG.Add(1)
    
    var err error
    client, err = rpc.DialHTTP(TYPE, PORT)
    if err != nil {
	panic(err)
    }
    err = client.Call("Receiver.Connect", &struct{}{}, &Token)
    if err != nil {
     log.Fatal(err) 
    }
    go UserIn()
    go userOut()
    
    waitG.Wait()
    fmt.Print(MESS_DISC)
  
}